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

package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

var tpl = `package {{.Package}}

import (
	"io/ioutil"
	"net/http"

	"code.uber.internal/example/example-gateway"
	{{.DownstreamService}}Client "code.uber.internal/example/example-gateway/clients/{{.DownstreamService}}"

	"github.com/pkg/errors"
)

func Handle{{.MyHandler | Title}}Request(
	inc *gateway.IncomingMessage,
	gateway *gateway.EdgeGateway,
) {
	rawBody, ok := inc.ReadAll()
	if !ok {
		return
	}

	var body {{.DownstreamService}}Client.{{.MyHandler | Title}}
	if ok := inc.UnmarshalBody(&body, rawBody); !ok {
		return
	}

	h := make(http.Header)
	h.Set("x-uber-uuid", inc.Header.Get("x-uber-uuid"))

	clientBody := convertToClient(&body)
	clientResp, err := g.Clients.{{.DownstreamService}}.{{.DownstreamMethod | Title}}(&body, h)
	if err != nil {
		gateway.Logger.Error("Could not make client request",
			zap.String("error", err.Error()),
		)		
		inc.SendError(500, errors.Wrap(err, "Could not make client request:"))
		return
	}

	defer func() {
		if err := clientResp.Body.Close(); err != nil {
			inc.SendError(500, errors.Wrap(err, "Could not close client response body:"))
			return
		}
	}()
	b, err := ioutil.ReadAll(clientResp.Body)
	if err != nil {
		inc.SendError(500, errors.Wrap(err, "Could not read client response body:"))
		return
	}

	if !isOKResponse(clientResp.StatusCode, []int{200, 202, 204}) {
		inc.SendErrorString(clientResp.StatusCode, string(b))
		return
	}

	// TODO(sindelar): Apply response filtering and translation.
	inc.CopyJSON(res.Res.StatusCode, res.Res.Body)
}

func convertToClient(
	body *{{.MyHandler | Title}},
) *{{.DownstreamService}}Client.{{.DownstreamMethod | Title}} {
	clientBody := &{{.DownstreamService}}Client.{{.DownstreamMethod | Title}}
    // TODO(sindelar): Add field mappings here.
    return body
	}
}
`

func main() {
	funcMap := template.FuncMap{
		"Title": strings.Title,
	}

	tt := template.Must(template.New("Handler").Funcs(funcMap).Parse(tpl))
    prefix := os.Args[1]

    // Iterate over all passed in endpoints.
	for i := 2; i < len(os.Args); i++ {
        endpoint := os.Args[i]
        endpointDir := prefix + string(os.PathSeparator) + strings.ToLower(endpoint)
        os.Mkdir(endpointDir, 0755)
        // PLACEHOLDER, REPLACE WITH HANDLERS FROM GENERATED CODE.
        // Read all handlers for an endpoint and generate each.
		handlers := []string{"foo"}
 		for j := 0; j < len(handlers); j++ {
			dest := endpointDir + string(os.PathSeparator) + strings.ToLower(handlers[j]) + "_handler.go"
			file, err := os.Create(dest)
			if err != nil {
				fmt.Printf("Could not create %s: %s (skip)\n", dest, err)
				continue
			}
			// TODO(sindelar): Use an endpoint to client map.
			downstreamService := endpoint
			downstreamMethod := os.Args[i]

			vals := map[string]string{
				"MyHandler": handlers[j],
				"Package": endpoint,
				"DownstreamService": downstreamService,
				"DownstreamMethod": downstreamMethod,
			}
			tt.Execute(file, vals)

			file.Close()
		}
	}
}
