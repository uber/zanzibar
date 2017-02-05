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
	"go/build"
	"os"
	"strings"
	"text/template"
	"path/filepath"
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

var clientResponse = "{\"statusCode\":200}"
var endpointPath = "/googlenow/add-credentials"
var endpointHttpMethod = "POST"
var clientPath = "/add-credentials"
var clientHttpMethod = "POST"

func main() {
	funcMap := template.FuncMap{
		"Title": strings.Title,
	}

	// Hacky way to find the template directory, is there a better way to do this?
	importPath := "github.com/uber/zanzibar/lib/gencode" // modify to match import path of main
	p, err := build.Default.Import(importPath, "", build.FindOnly)
	if err != nil {
		panic(fmt.Sprintf("Could not create build path for endpoint generation: %s", err))
		// handle error
	}
	templatePath := filepath.Join(p.Dir, "templates/endpoint_template")

	//templatePath, _ := filepath.Abs("lib/gencode/templates/endpoint_template")
	//var templatePaths [1]string
	//templatePaths[0] = templatePath

    //b, err := ioutil.ReadFile(templatePath) // just pass the file name
    //if err != nil {
    //    fmt.Print(err)
    //}
    //str := string(b) // convert content to a 'string'
    //fmt.Println(str) // print the content as a 'string'

	tt := template.Must(template.New("Handler").Funcs(funcMap).ParseFiles(templatePath))
    prefix := os.Args[1]

	fmt.Printf("Template %s\n", tt)

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
			tt.ExecuteTemplate(file, "Handler", vals)

			file.Close()

            // Build a test file with benchmarking to validate the endpoint.
            // In the future, refactor this into a test runner instead of generating test
			// files for each endpoint.
        	testCase := template.Must(template.New("HandlerTests").Funcs(funcMap).Parse(testTemplate))
			fmt.Printf("Template %s\n", testCase)
			dest = endpointDir + string(os.PathSeparator) + strings.ToLower(handlers[j]) + "_handler_test.go"
			file, err = os.Create(dest)
			if err != nil {
				fmt.Printf("Could not create %s: %s (skip)\n", dest, err)
				continue
			}
			vals = map[string]string{
				"MyHandler": handlers[j],
				"Package": endpoint,
				"DownstreamService": downstreamService,
				"DownstreamMethod": downstreamMethod,
                "EndpointPath": endpointPath,
                "EndpointHttpMethod": endpointHttpMethod,
                "ClientPath": clientPath,
                "ClientHttpMethod": clientHttpMethod,
                "ClientResponse": clientResponse,
			}
            testCase.Execute(file, vals)

			file.Close()
		}
	}
}

var testTemplate = `package {{.MyHandler}}_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	assert "github.com/stretchr/testify/assert"
	config "github.com/uber/zanzibar/examples/example-gateway/config"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testBackend "github.com/uber/zanzibar/test/lib/test_backend"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
)

var benchBytes = []byte({{.EndpointInput}}})

type testCase struct {
	Counter      int
	IsBench      bool
	Backend      *testBackend.TestBackend
	TestGateway  *testGateway.TestGateway
	BenchGateway *benchGateway.BenchGateway
}

func (testCase *testCase) Close() {
	testCase.Backend.Close()

	if testCase.IsBench {
		testCase.BenchGateway.Close()
	} else {
		testCase.TestGateway.Close()
	}
}

func newTestCase(t *testing.T, isBench bool) (*testCase, error) {
	testCase := &testCase{
		IsBench: isBench,
	}

	testCase.Backend = testBackend.CreateBackend(0)
	err := testCase.Backend.Bootstrap()
	if err != nil {
		return nil, err
	}

    {{.DownstreamMethod}} := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte({{.ClientResponse}})); err != nil {
			t.Fatal("can't write fake response")
		}
		testCase.Counter++
	}
	testCase.Backend.HandleFunc({{.ClientHttpMethod}}, {{.ClientPath}}, {{.DownstreamMethod}})

	config := &config.Config{}
	config.Clients.GoogleNow.IP = "127.0.0.1"
	config.Clients.GoogleNow.Port = testCase.Backend.RealPort

	if testCase.IsBench {
		gateway, err := benchGateway.CreateGateway(config)
		if err != nil {
			return nil, err
		}
		testCase.BenchGateway = gateway
	} else {
		gateway, err := testGateway.CreateGateway(t, config, nil)
		if err != nil {
			return nil, err
		}
		testCase.TestGateway = gateway
	}
	return testCase, nil
}

func Benchmark{{.MyEndpoint}}{{.MyHandler | Title}}(b *testing.B) {
	testCase, err := newTestCase(nil, true)
	if err != nil {
		b.Error("got bootstrap err: " + err.Error())
		return
	}

	b.ResetTimer()

	// b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := testCase.BenchGateway.MakeRequest(
				"POST", "/googlenow/add-credentials",
				bytes.NewReader(benchBytes),
			)
			if err != nil {
				b.Error("got http error: " + err.Error())
				break
			}
			if res.Status != "200 OK" {
				b.Error("got bad status error: " + res.Status)
				break
			}

			_, err = ioutil.ReadAll(res.Body)
			if err != nil {
				b.Error("could not write response: " + res.Status)
				break
			}
		}
	})

	b.StopTimer()
	testCase.Close()
	b.StartTimer()
}

func TestAddCredentials(t *testing.T) {
	testCase, err := newTestCase(t, false)
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer testCase.Close()

	assert.NotNil(t, testCase.TestGateway, "gateway exists")

	res, err := testCase.TestGateway.MakeRequest(
		{{.EndpointHttpMethod}}, {{.EndpointPath}}, bytes.NewReader(benchBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, testCase.Counter)
}
`
