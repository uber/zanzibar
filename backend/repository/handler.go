// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package repository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Handler provides http handlers for modifying gateway configurations.
type Handler struct {
	Manager       *Manager
	DiffCreator   DiffCreator
	gatewayHeader string
	logger        Logger
}

// DiffCreator creates and lands a diff from changes of local repository.
type DiffCreator interface {
	// NewDiff return a URI for the diff or an error.
	NewDiff(r *Repository, request *DiffRequest) (string, error)
	// LandDiff lands a diff into a remote repository.
	LandDiff(r *Repository, diffURI string) error
}

// Logger is based on the interface of Zap logger.
type Logger interface {
	Info(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
}

// DiffRequest contains all information to create a diff.
type DiffRequest struct {
	BranchName    string
	CommitMessage string
	Reviewers     []string
}

// NewHandler constructs a new Handler.
func NewHandler(m *Manager, dc DiffCreator, gatewayHeader string, logger Logger) *Handler {
	return &Handler{
		Manager:       m,
		DiffCreator:   dc,
		gatewayHeader: gatewayHeader,
		logger:        logger,
	}
}

// NewHTTPRouter constructs the endpoints for http server.
func (h *Handler) NewHTTPRouter() *httprouter.Router {
	r := httprouter.New()
	// TODO(zw): Add more endpoints and tests.
	r.GET("/gateways", h.GatewayAll)
	r.POST("/create-diff", h.CreateDiff)
	return r
}

// GatewayAll returns configurations of all edge gateways.
func (h *Handler) GatewayAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cfgMap := make(map[string]*Config, len(h.Manager.RepoMap))
	for _, repo := range h.Manager.RepoMap {
		cfg, err := repo.GatewayConfig()
		if err != nil {
			h.WriteJSON(w, http.StatusInternalServerError, err)
			return
		}
		cfgMap[cfg.ID] = cfg
	}
	h.WriteJSON(w, http.StatusOK, cfgMap)
}

type createDiffResponse struct {
	DiffURI string `json:"diff_uri"`
}

// CreateDiff creates a diff for updates for thrift files, clients and endpoints.
func (h *Handler) CreateDiff(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := &UpdateRequest{}
	b, err := UnmarshalJSONBody(r, req)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.Wrap(err, "Failed to unmarshal body for createing a diff"))
		return
	}
	h.logger.Info("Creating a diff.",
		zap.String("request", string(b)),
	)
	gatewayConfig, err := h.GatewayConfig(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusNotFound, err)
		return
	}
	repo, err := h.Manager.NewRuntimeRepository(gatewayConfig.ID)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, errors.Wrap(err, "failed to create temp runtime dir"))
		return
	}
	if err := h.Manager.UpdateAll(repo, gatewayConfig.ClientConfigDir, gatewayConfig.EndpointConfigDir, req); err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	if _, err := repo.LatestGatewayConfig(); err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, errors.Wrap(err, "invalid gateway config after upadte"))
		return
	}
	diffReq := &DiffRequest{
		BranchName:    strings.Join([]string{"batch", "update", currentTime()}, "_"),
		CommitMessage: fmt.Sprintf("[%s] Batch update", gatewayConfig.ID),
	}
	h.logger.Info("Creating diff...",
		zap.String("localDir", repo.LocalDir()),
		zap.String("remote", repo.Remote()),
		zap.String("diffRequest", fmt.Sprintf("%+v", diffReq)),
	)
	resp, err := h.DiffCreator.NewDiff(repo, diffReq)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, resp)
		return
	}
	h.WriteJSON(w, http.StatusOK, &createDiffResponse{DiffURI: resp})
}

// WriteJSON writes the JSON response.
func (h *Handler) WriteJSON(w http.ResponseWriter, code int, response interface{}) {
	if code != http.StatusOK {
		h.logger.Error("Error Response.",
			zap.String("response", fmt.Sprintf("%+v", response)),
		)
	}
	b, err := json.Marshal(response)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(code)
	_, err = w.Write(b)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
	}
}

// WriteErrorResponse writes an error HTTP response.
func (h *Handler) WriteErrorResponse(w http.ResponseWriter, code int, err error) {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	var errStack string
	if causeErr, ok := errors.Cause(err).(stackTracer); ok {
		errStack = fmt.Sprintf("%+v", causeErr.StackTrace())
	} else {
		errStack = "unknown error stack"
	}
	h.logger.Error("Error Response.",
		zap.Int("Status", code),
		zap.String("errorString", fmt.Sprintf("%s", err)),
		zap.String("errorStack", errStack),
	)
	w.WriteHeader(code)
	if _, e := w.Write([]byte(fmt.Sprintf("{\"error\": %q}", err.Error()))); e != nil {
		h.logger.Error("Failed to write error response.",
			zap.Int("statusCode", code),
			zap.String("errorString", fmt.Sprintf("%s", err)),
			zap.String("errorStack", errStack),
			zap.String("writeError", fmt.Sprintf("%+v", e)),
		)
	}
}

// GatewayConfig returns configuration for a http request.
func (h *Handler) GatewayConfig(r *http.Request) (*Config, error) {
	id := r.Header.Get(h.gatewayHeader)
	repo, ok := h.Manager.RepoMap[id]
	if !ok {
		return nil, errors.Errorf("gateway with ID %q from header %q is not found", id, h.gatewayHeader)
	}
	gatewayConfig, err := repo.GatewayConfig()
	if err != nil {
		return nil, err
	}
	return gatewayConfig, nil
}

// UnmarshalJSONBody returns the parsed JSON body.
func UnmarshalJSONBody(r *http.Request, out interface{}) ([]byte, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read request body")
	}
	if err := json.Unmarshal(b, out); err != nil {
		return nil, errors.Wrapf(err, "body %s", b)
	}
	return b, nil
}

func currentTime() string {
	return time.Now().Format("2006-01-02T15-04-05Z07-00")
}
