// Copyright (c) 2018 Uber Technologies, Inc.
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
	reqerr "github.com/uber/zanzibar/codegen/errors"
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
	// UpdateDiff return a URI for the diff or an error.
	UpdateDiff(r *Repository, request *DiffRequest) (string, error)
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
	DiffID        string
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
	r.PanicHandler = h.handlePanic

	// TODO(zw): Add more endpoints and tests.
	r.GET("/gateways", h.GatewayAll)
	r.GET("/gateway/:id", h.GatewayByID)
	r.GET("/endpoint/:id", h.EndpointByID)
	r.GET("/clients", h.ClientAll)
	r.GET("/client/:name", h.ClientByName)
	r.GET("/middlewares", h.MiddlewareAll)
	r.GET("/idl-registry-list", h.IDLRegistryList)
	r.GET("/idl-registry/*path", h.IDLRegistryFile)
	r.GET("/idl-registry-service/*path", h.IDLRegistryThriftService)
	r.GET("/thrift-services", h.ThriftServicesAll)
	r.GET("/thrift-service/*path", h.ThriftServicesByPath)
	r.GET("/thrift-list", h.ThriftList)
	r.GET("/thrift-file/*path", h.ThriftFile)
	r.GET("/thrift-file-compiled/*path", h.CompiledThrift)
	r.GET("/thrift-method-parsed/*path", h.MethodFromThriftCode)
	r.POST("/validate-updates", h.ValidateUpdates)
	r.POST("/create-diff", h.GenerateDiff)
	r.POST("/land-diff", h.LandDiff)
	r.POST("/thrift-file-parsed", h.ThriftModuleToCode)
	r.POST("/thrift-file-code/*path", h.CodeThrift)
	return r
}

func (h *Handler) handlePanic(
	w http.ResponseWriter, r *http.Request, v interface{},
) {
	err, ok := v.(error)
	if !ok {
		err = errors.Errorf("backend handler panic: %v", v)
	}
	_, ok = err.(fmt.Formatter)
	if !ok {
		err = errors.Wrap(err, "wrapped")
	}

	h.logger.Error(
		"A backed request handler panicked",
		zap.Error(err),
		zap.String("pathname", r.URL.RequestURI()),
	)

	h.WriteJSON(w,
		http.StatusInternalServerError,
		map[string]string{"message": err.Error()},
	)
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

// GatewayByID returns configurations of a edge gateway.
func (h *Handler) GatewayByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	repo, ok := h.Manager.RepoMap[id]
	if !ok {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.Errorf("Gateway %q is not found.", id))
		return
	}
	cfg, err := repo.GatewayConfig()
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, err)
		return
	}
	h.WriteJSON(w, http.StatusOK, cfg)
}

// EndpointByID returns the configuration of a given endpoint.
func (h *Handler) EndpointByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	gatewayConfig, err := h.GatewayConfig(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusNotFound, err)
		return
	}
	id := ps.ByName("id")
	endpoint, ok := gatewayConfig.Endpoints[id]
	if !ok {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.Errorf("Endpoint %q is not found.", id))
		return
	}
	h.WriteJSON(w, http.StatusOK, endpoint)
}

// ClientAll returns configurations for all clients.
func (h *Handler) ClientAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	gatewayConfig, err := h.GatewayConfig(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusNotFound, err)
		return
	}
	h.WriteJSON(w, http.StatusOK, gatewayConfig.Clients)
}

// ClientByName returns the client configuration by name.
func (h *Handler) ClientByName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	gatewayConfig, err := h.GatewayConfig(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusNotFound, err)
		return
	}
	name := ps.ByName("name")
	client, ok := gatewayConfig.Clients[name]
	if !ok {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.Errorf("Client %q is not found.", name))
		return
	}
	h.WriteJSON(w, http.StatusOK, client)
}

// MiddlewareAll returns configurations for all middlewares.
func (h *Handler) MiddlewareAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	gatewayConfig, err := h.GatewayConfig(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusNotFound, err)
		return
	}
	h.WriteJSON(w, http.StatusOK, gatewayConfig.Middlewares)
}

// ThriftServicesAll returns thrift services available in a gateway.
func (h *Handler) ThriftServicesAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	gatewayConfig, err := h.GatewayConfig(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusNotFound, err)
		return
	}
	h.WriteJSON(w, http.StatusOK, gatewayConfig.ThriftServices)
}

// ThriftServicesByPath returns sevices in a thrift file of a gateway.
func (h *Handler) ThriftServicesByPath(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	gatewayConfig, err := h.GatewayConfig(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusNotFound, err)
		return
	}
	path := strings.TrimLeft(ps.ByName("path"), "/")
	h.logger.Info("Thrift services by path.", zap.String("file", path))
	thrift, ok := gatewayConfig.ThriftServices[path]
	if ok {
		h.WriteJSON(w, http.StatusOK, thrift)
		return
	}
	// Thrift file not found in gateway. Try to find it in IDL-registry then.
	thrift, err = h.Manager.IDLThriftService(path)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.Wrapf(err, "file %q is not found in gateway and IDL-registry", path))
		return
	}
	h.WriteJSON(w, http.StatusOK, thrift)
}

// IDLRegistryList returns the full list of files in IDL-registry.
func (h *Handler) IDLRegistryList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	metaMap, err := h.Manager.IDLRegistry.ThriftAll()
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	h.WriteJSON(w, http.StatusOK, metaMap)
}

// IDLRegistryFile returns the content and meta data of a file in IDL-registry.
func (h *Handler) IDLRegistryFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	path := strings.TrimLeft(ps.ByName("path"), "/")
	meta, err := h.Manager.IDLRegistry.ThriftMeta(path, true)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	h.WriteJSON(w, http.StatusOK, meta)
}

// IDLRegistryThriftService returns the services in a thrift file in IDL-registry.
func (h *Handler) IDLRegistryThriftService(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	path := strings.TrimLeft(ps.ByName("path"), "/")
	thriftServices, err := h.Manager.IDLThriftService(path)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	h.WriteJSON(w, http.StatusOK, thriftServices)
}

// ThriftList returns the full list of thrift files in a gateway.
func (h *Handler) ThriftList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := r.Header.Get(h.gatewayHeader)
	list, err := h.Manager.ThriftList(id)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	h.WriteJSON(w, http.StatusOK, list)
}

// ThriftFile returns the content and meta data of a file in a gateway.
func (h *Handler) ThriftFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := r.Header.Get(h.gatewayHeader)
	path := strings.TrimLeft(ps.ByName("path"), "/")
	meta, err := h.Manager.ThriftFile(id, path)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	h.WriteJSON(w, http.StatusOK, meta)
}

// CompiledThrift returns the content and meta data of a file in a gateway.
func (h *Handler) CompiledThrift(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := r.Header.Get(h.gatewayHeader)
	path := strings.TrimLeft(ps.ByName("path"), "/")
	module, err := h.Manager.CompileThriftFile(id, path)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	h.WriteJSON(w, http.StatusOK, module)
}

type methodFromThriftCodeResponse struct {
	Functions []string `json:"functions"`
}

// MethodFromThriftCode returns a list of method names for a given thrift file path
func (h *Handler) MethodFromThriftCode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := r.Header.Get(h.gatewayHeader)
	path := strings.TrimLeft(ps.ByName("path"), "/")
	module, err := h.Manager.CompileThriftFile(id, path)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.Wrap(err, "Failed to compile thrift file from the given path"))
		return
	}

	var functionNames []string
	for _, serviceSpec := range module.Services {
		for _, functionSpec := range serviceSpec.Functions {
			functionNames = append(functionNames, serviceSpec.Name+"::"+functionSpec.Name)
		}
	}
	h.WriteJSON(w, http.StatusOK, &methodFromThriftCodeResponse{Functions: functionNames})
}

type rawCodeRequest struct {
	Content string `json:"content"`
}

// CodeThrift takes raw code, validates and returns the parsed Module version
func (h *Handler) CodeThrift(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := &rawCodeRequest{}
	b, err := UnmarshalJSONBody(r, req)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.Wrap(err, "Failed to unmarshal body for converting raw code"))
		return
	}
	h.logger.Info("Validating and parsing thrift code", zap.String("request", string(b)))
	id := r.Header.Get(h.gatewayHeader)
	path := strings.TrimLeft(ps.ByName("path"), "/")
	module, err := h.Manager.CodeThriftFile(req.Content, id, path)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	h.WriteJSON(w, http.StatusOK, module)
}

// ValidateUpdates validates the update requests for thrift files, clients and endpoints.
func (h *Handler) ValidateUpdates(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := &UpdateRequest{}
	b, err := UnmarshalJSONBody(r, req)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.Wrap(err, "Failed to unmarshal body for validating updates"))
		return
	}
	h.logger.Info("Validating update request.", zap.String("request", string(b)))
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
	h.WriteJSON(w, http.StatusOK, map[string]string{
		"Status": "OK",
	})
}

type thriftToCodeResponse struct {
	Content string `json:"content"`
}

// ThriftModuleToCode takes a module structure and converts it to code
func (h *Handler) ThriftModuleToCode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := &Module{}
	b, err := UnmarshalJSONBody(r, req)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.Wrap(err, "Failed to unmarshal body for converting a thrift module"))
		return
	}
	h.logger.Info("converting to a thrift module.",
		zap.String("request", string(b)),
	)

	code := req.Code()
	h.WriteJSON(w, http.StatusOK, &thriftToCodeResponse{Content: code})
}

type generateDiffResponse struct {
	DiffURI string `json:"diff_uri"`
}

// GenerateDiff generates diff for thrift files, client and endpoints.
// optionally it takes a diff id to updating an existing diff
func (h *Handler) GenerateDiff(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var (
		resp string
		err  error
		req  *UpdateRequest
	)
	req = &UpdateRequest{}
	b, err := UnmarshalJSONBody(r, req)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.Wrap(err, "Failed to unmarshal body for createing a diff"))
		return
	}
	h.logger.Info("Generating a diff.",
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
	if err := h.Manager.Validate(repo, req); err != nil {
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
	h.logger.Info("Generating diff...",
		zap.String("localDir", repo.LocalDir()),
		zap.String("remote", repo.Remote()),
		zap.String("diffRequest", fmt.Sprintf("%+v", diffReq)),
	)
	if req.DIffID == nil {
		resp, err = h.DiffCreator.NewDiff(repo, diffReq)
	} else {
		diffReq.DiffID = *req.DIffID
		resp, err = h.DiffCreator.UpdateDiff(repo, diffReq)
	}
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, resp)
		return
	}
	h.WriteJSON(w, http.StatusOK, &generateDiffResponse{DiffURI: resp})
}

// LandDiff lands a diff.
func (h *Handler) LandDiff(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	type LandDiffRequest struct {
		DiffURL string `json:"diff_url"`
	}
	request := &LandDiffRequest{}
	b, err := UnmarshalJSONBody(r, request)
	h.logger.Info("Landing a diff.",
		zap.String("request", string(b)),
	)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.Wrap(err, "Failed to unmarshal body for landing diff"))
		return
	}
	gatewayConfig, err := h.GatewayConfig(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusNotFound, err)
		return
	}
	repo, err := h.Manager.NewRuntimeRepository(gatewayConfig.ID)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, errors.Wrap(err, "failed to create temp runtime dirs"))
		return
	}
	h.logger.Info("Landing diff...",
		zap.String("localDir", repo.LocalDir()),
		zap.String("remote", repo.Remote()),
		zap.String("diffURL", request.DiffURL),
	)
	err = h.DiffCreator.LandDiff(repo, request.DiffURL)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	h.WriteJSON(w, http.StatusOK, map[string]string{
		"Status": "OK",
	})
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
	errJSON := fmt.Sprintf("{\"error\": %q}", err.Error())
	causeErr := errors.Cause(err)
	// In case we know the real cause of the error.
	if e, ok := causeErr.(*reqerr.RequestError); ok {
		code = http.StatusBadRequest
		errJSON = e.JSON()
		causeErr = e.Cause
	}

	var errStack string
	if e, ok := causeErr.(stackTracer); ok {
		errStack = fmt.Sprintf("%+v", e.StackTrace())
	} else {
		errStack = "unknown error stack"
	}
	h.logger.Error("Error Response.",
		zap.Int("Status", code),
		zap.String("errorJSON", fmt.Sprintf("%s", errJSON)),
		zap.String("errorRaw", fmt.Sprintf("%s", err)),
		zap.String("errorStack", errStack),
	)
	w.WriteHeader(code)
	if _, e := w.Write([]byte(errJSON)); e != nil {
		h.logger.Error("Failed to write error response.",
			zap.Int("statusCode", code),
			zap.String("errorString", fmt.Sprintf("%s", errJSON)),
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
