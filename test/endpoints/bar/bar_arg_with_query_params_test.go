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

package bar_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

var barResponseBytes = `{
	"stringField":"stringValue",
	"intWithRange":0,
	"intWithoutRange":0,
	"mapIntWithRange":{},
	"mapIntWithoutRange":{},
	"binaryField":"d29ybGQ="
}`

var barResponseBytesRecursive = `{
	"stringField":"new str val",
	"intWithRange":4,
	"intWithoutRange":6,
	"mapIntWithRange":{},
	"mapIntWithoutRange":{},
	"binaryField":"aGV5IHdvcmxk",
	"nextResponse":` + barResponseBytes + `
}`

func TestBarWithQueryParamsCall(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t,
				"name=foo&userUUID=bar",
				r.URL.RawQuery,
			)

			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithQueryParams?name=foo&userUUID=bar",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(barResponseBytes))
}

func TestBarWithQueryParamsCallWithRecursiveResponse(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t,
				"name=foo&userUUID=bar",
				r.URL.RawQuery,
			)

			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytesRecursive)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithQueryParams?name=foo&userUUID=bar",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(barResponseBytesRecursive))
}

func TestBarWithQueryParamsCallWithMalformedQuery(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t,
				"name=foo&userUUID=bar",
				r.URL.RawQuery,
			)

			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithQueryParams?%gh&%ij",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "400 Bad Request", res.Status)
	assert.Equal(t, 0, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(`{
		"error":"Could not parse query string"
	}`))

	logLines := gateway.Logs("warn", "Got request with invalid query string")
	assert.Equal(t, len(logLines), 1)

	line := logLines[0]
	assert.Equal(t, line["error"].(string), `invalid URL escape "%gh"`)
}

func TestBarWithManyQueryParamsCall(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithManyQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t,
				"aBoolean=true&aFloat64=5.1&aInt16=48&aInt32=12&"+
					"aInt64=4&aInt8=24&aStr=foo&anOptBool=false&"+
					"anOptFloat64=-0.4&anOptInt16=-100&anOptInt32=-10&"+
					"anOptInt64=-1&anOptInt8=-50&anOptStr=bar",
				r.URL.RawQuery,
			)

			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithManyQueryParams?"+
			"aStr=foo&anOptStr=bar&aBool=true&anOptBool=false&"+
			"aInt8=24&anOptInt8=-50&aInt16=48&anOptInt16=-100&"+
			"aInt32=12&anOptInt32=-10&aInt64=4&anOptInt64=-1&"+
			"aFloat64=5.1&anOptFloat64=-0.4",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(barResponseBytes))
}

func TestBarManyQueryParamsWithInvalidBool(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithManyQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithManyQueryParams?"+
			"aStr=foo&anOptStr=bar&aBool=t&anOptBool=false&"+
			"aInt8=24&anOptInt8=-50&aInt16=48&anOptInt16=-100&"+
			"aInt32=12&anOptInt32=-10&aInt64=4&anOptInt64=-1&"+
			"aFloat64=5.1&anOptFloat64=-0.4",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "400 Bad Request", res.Status)
	assert.Equal(t, 0, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(`{
		"error":"Could not parse query string"
	}`))

	logLines := gateway.Logs(
		"warn", "Got request with invalid query string types",
	)
	assert.Equal(t, 1, len(logLines))

	line := logLines[0]
	assert.Equal(t,
		"strconv.ParseBool: parsing \"t\": invalid syntax",
		line["error"].(string),
	)
}

func TestBarManyQueryParamsWithInvalidInt8(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithManyQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithManyQueryParams?"+
			"aStr=foo&anOptStr=bar&aBool=true&anOptBool=false&"+
			"aInt8=wat&anOptInt8=-50&aInt16=48&anOptInt16=-100&"+
			"aInt32=12&anOptInt32=-10&aInt64=4&anOptInt64=-1&"+
			"aFloat64=5.1&anOptFloat64=-0.4",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "400 Bad Request", res.Status)
	assert.Equal(t, 0, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(`{
		"error":"Could not parse query string"
	}`))

	logLines := gateway.Logs(
		"warn", "Got request with invalid query string types",
	)
	assert.Equal(t, 1, len(logLines))

	line := logLines[0]
	assert.Equal(t,
		"strconv.ParseInt: parsing \"wat\": invalid syntax",
		line["error"].(string),
	)
}

func TestBarManyQueryParamsWithInvalidInt16(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithManyQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithManyQueryParams?"+
			"aStr=foo&anOptStr=bar&aBool=true&anOptBool=false&"+
			"aInt8=24&anOptInt8=-50&aInt16=wat&anOptInt16=-100&"+
			"aInt32=12&anOptInt32=-10&aInt64=4&anOptInt64=-1&"+
			"aFloat64=5.1&anOptFloat64=-0.4",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "400 Bad Request", res.Status)
	assert.Equal(t, 0, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(`{
		"error":"Could not parse query string"
	}`))

	logLines := gateway.Logs(
		"warn", "Got request with invalid query string types",
	)
	assert.Equal(t, 1, len(logLines))

	line := logLines[0]
	assert.Equal(t,
		"strconv.ParseInt: parsing \"wat\": invalid syntax",
		line["error"].(string),
	)
}

func TestBarManyQueryParamsWithInvalidInt32(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithManyQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithManyQueryParams?"+
			"aStr=foo&anOptStr=bar&aBool=true&anOptBool=false&"+
			"aInt8=24&anOptInt8=-50&aInt16=48&anOptInt16=-100&"+
			"aInt32=wat&anOptInt32=-10&aInt64=4&anOptInt64=-1&"+
			"aFloat64=5.1&anOptFloat64=-0.4",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "400 Bad Request", res.Status)
	assert.Equal(t, 0, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(`{
		"error":"Could not parse query string"
	}`))

	logLines := gateway.Logs(
		"warn", "Got request with invalid query string types",
	)
	assert.Equal(t, 1, len(logLines))

	line := logLines[0]
	assert.Equal(t,
		"strconv.ParseInt: parsing \"wat\": invalid syntax",
		line["error"].(string),
	)
}

func TestBarManyQueryParamsWithInvalidInt64(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithManyQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithManyQueryParams?"+
			"aStr=foo&anOptStr=bar&aBool=true&anOptBool=false&"+
			"aInt8=24&anOptInt8=-50&aInt16=48&anOptInt16=-100&"+
			"aInt32=12&anOptInt32=-10&aInt64=wat&anOptInt64=-1&"+
			"aFloat64=5.1&anOptFloat64=-0.4",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "400 Bad Request", res.Status)
	assert.Equal(t, 0, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(`{
		"error":"Could not parse query string"
	}`))

	logLines := gateway.Logs(
		"warn", "Got request with invalid query string types",
	)
	assert.Equal(t, 1, len(logLines))

	line := logLines[0]
	assert.Equal(t,
		"strconv.ParseInt: parsing \"wat\": invalid syntax",
		line["error"].(string),
	)
}

func TestBarManyQueryParamsWithInvalidFloat64(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithManyQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithManyQueryParams?"+
			"aStr=foo&anOptStr=bar&aBool=true&anOptBool=false&"+
			"aInt8=24&anOptInt8=-50&aInt16=48&anOptInt16=-100&"+
			"aInt32=12&anOptInt32=-10&aInt64=4&anOptInt64=-1&"+
			"aFloat64=wat&anOptFloat64=-0.4",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "400 Bad Request", res.Status)
	assert.Equal(t, 0, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(`{
		"error":"Could not parse query string"
	}`))

	logLines := gateway.Logs(
		"warn", "Got request with invalid query string types",
	)
	assert.Equal(t, 1, len(logLines))

	line := logLines[0]
	assert.Equal(t,
		"strconv.ParseFloat: parsing \"wat\": invalid syntax",
		line["error"].(string),
	)
}

func TestBarWithQueryHeaders(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar/argWithQueryHeader",
		func(w http.ResponseWriter, r *http.Request) {
			bytes, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t,
				`{"userUUID":"a-uuid"}`,
				string(bytes),
			)

			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithQueryHeader",
		map[string]string{
			"x-uuid": "a-uuid",
		}, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(barResponseBytes))
}

func TestBarWithManyQueryParamsRequiredCall(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar/argWithManyQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithManyQueryParams",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "400 Bad Request", res.Status)
	assert.Equal(t, 0, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	logs := gateway.AllLogs()

	assert.Equal(t, 4, len(logs))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request"]))
	assert.Equal(t, 1, len(logs["Started ExampleGateway"]))
	assert.Equal(t, 1, len(logs["Got request with missing query string value"]))
	assert.Equal(t, 1, len(logs["Sending error for endpoint request"]))

	assert.Equal(t,
		"aStr",
		logs["Got request with missing query string value"][0]["expectedKey"],
	)
}

func TestBarWithManyQueryParamsOptionalCall(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithManyQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t,
				"aBoolean=true&aFloat64=5.1&aInt16=48&aInt32=12&"+
					"aInt64=4&aInt8=24&aStr=foo&anOptBool=false&"+
					"anOptInt8=-50&anOptStr=bar",
				r.URL.RawQuery,
			)

			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithManyQueryParams?"+
			"aStr=foo&anOptStr=bar&aBool=true&anOptBool=false&"+
			"aInt8=24&anOptInt8=-50&aInt16=48&"+
			"aInt32=12&aInt64=4&aFloat64=5.1",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(barResponseBytes))
}

func TestBarWithNestedQueryParams(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithNestedQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t,
				"request.authUUID=auth-uuid&request.myuuid=auth-uuid2&"+
					"request.name=a-name&request.userUUID=a-uuid",
				r.URL.RawQuery,
			)

			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithNestedQueryParams?"+
			"request.name=a-name&request.userUUID=a-uuid",
		map[string]string{
			"x-uuid":  "auth-uuid",
			"x-uuid2": "auth-uuid2",
		}, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(barResponseBytes))
}

func TestBarWithNestedQueryParamsWithOpts(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithNestedQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t,
				"opt.name=b-name&opt.userUUID=b-uuid&"+
					"request.authUUID=auth-uuid&request.myuuid=auth-uuid2&"+
					"request.name=a-name&request.userUUID=a-uuid",
				r.URL.RawQuery,
			)

			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithNestedQueryParams?"+
			"request.name=a-name&request.userUUID=a-uuid&"+
			"opt.name=b-name&opt.userUUID=b-uuid",
		map[string]string{
			"x-uuid":  "auth-uuid",
			"x-uuid2": "auth-uuid2",
		}, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(barResponseBytes))
}

func TestBarWithNestedQueryParamsWithoutHeaders(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/argWithNestedQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t,
				"request.name=a-name&request.userUUID=a-uuid",
				r.URL.RawQuery,
			)

			w.WriteHeader(200)
			if _, err := w.Write([]byte(barResponseBytes)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"GET",
		"/bar/argWithNestedQueryParams?"+
			"request.name=a-name&request.userUUID=a-uuid",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes), compactStr(barResponseBytes))
}
