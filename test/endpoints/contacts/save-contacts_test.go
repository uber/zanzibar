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

package save_contacts_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/examples/example-gateway/endpoints/contacts"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
	"github.com/uber/zanzibar/test/lib/test_gateway"
)

var benchBytes = []byte("{\"contacts\":[{\"fragments\":[{\"type\":\"message\",\"text\":\"foobarbaz\"}],\"attributes\":{\"firstName\":\"steve\",\"lastName\":\"stevenson\",\"hasPhoto\":true,\"numFields\":10,\"timesContacted\":5,\"lastTimeContacted\":0,\"isStarred\":false,\"hasCustomRingtone\":false,\"isSendToVoicemail\":false,\"hasThumbnail\":false,\"namePrefix\":\"\",\"nameSuffix\":\"\"}},{\"fragments\":[{\"type\":\"message\",\"text\":\"foobarbaz\"}],\"attributes\":{\"firstName\":\"steve\",\"lastName\":\"stevenson\",\"hasPhoto\":true,\"numFields\":10,\"timesContacted\":5,\"lastTimeContacted\":0,\"isStarred\":false,\"hasCustomRingtone\":false,\"isSendToVoicemail\":false,\"hasThumbnail\":false,\"namePrefix\":\"\",\"nameSuffix\":\"\"}},{\"fragments\":[],\"attributes\":{\"firstName\":\"steve\",\"lastName\":\"stevenson\",\"hasPhoto\":true,\"numFields\":10,\"timesContacted\":5,\"lastTimeContacted\":0,\"isStarred\":false,\"hasCustomRingtone\":false,\"isSendToVoicemail\":false,\"hasThumbnail\":false,\"namePrefix\":\"\",\"nameSuffix\":\"\"}},{\"fragments\":[],\"attributes\":{\"firstName\":\"steve\",\"lastName\":\"stevenson\",\"hasPhoto\":true,\"numFields\":10,\"timesContacted\":5,\"lastTimeContacted\":0,\"isStarred\":false,\"hasCustomRingtone\":false,\"isSendToVoicemail\":false,\"hasThumbnail\":false,\"namePrefix\":\"\",\"nameSuffix\":\"\"}}],\"appType\":\"MY_APP\"}")

func BenchmarkSaveContacts(b *testing.B) {
	gateway, err := benchGateway.CreateGateway(nil, &testGateway.Options{
		KnownHTTPBackends: []string{"contacts"},
	})
	if err != nil {
		b.Error("got bootstrap err: " + err.Error())
		return
	}

	gateway.HTTPBackends()["contacts"].HandleFunc(
		"POST", "/foo/contacts", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(202)
			_, _ = w.Write([]byte("{}"))
		},
	)

	b.ResetTimer()

	// b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := gateway.MakeRequest(
				"POST", "/contacts/foo/contacts", nil,
				bytes.NewReader(benchBytes),
			)
			if err != nil {
				b.Error("got http error: " + err.Error())
				break
			}
			if res.Status != "202 Accepted" {
				b.Error("got bad status error: " + res.Status)
				break
			}

			_, err = ioutil.ReadAll(res.Body)
			if err != nil {
				b.Error("could not write response: " + res.Status)
				break
			}
			_ = res.Body.Close()
		}
	})

	b.StopTimer()
	gateway.Close()
	b.StartTimer()
}

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(file)
}

func TestSaveContactsCall(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"contacts"},
		TestBinary: filepath.Join(
			getDirName(), "..", "..", "..",
			"examples", "example-gateway", "build", "main.go",
		),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["contacts"].HandleFunc(
		"POST", "/foo/contacts", func(w http.ResponseWriter, r *http.Request) {
			counter++
			w.WriteHeader(202)
			_, _ = w.Write([]byte("{}"))
		},
	)

	saveContacts := &contacts.SaveContactsRequest{
		Contacts: []*contacts.Contact{},
	}
	rawBody, _ := saveContacts.MarshalJSON()

	res, err := gateway.MakeRequest(
		"POST", "/contacts/foo/contacts", nil, bytes.NewReader(rawBody),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "202 Accepted", res.Status)
	assert.Equal(t, 1, counter)
}
