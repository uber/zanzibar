// Copyright (c) 2022 Uber Technologies, Inc.
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
	"testing"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/stretchr/testify/assert"
	endpointContacts "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/contacts/contacts"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

var benchBytes = []byte("{\"saveContactsRequest\": {\"userUUID\":\"foo\",\"contacts\":[{\"fragments\":[{\"type\":\"message\",\"text\":\"foobarbaz\"}],\"attributes\":{\"firstName\":\"steve\",\"lastName\":\"stevenson\",\"hasPhoto\":true,\"numFields\":10,\"timesContacted\":5,\"lastTimeContacted\":0,\"isStarred\":false,\"hasCustomRingtone\":false,\"isSendToVoicemail\":false,\"hasThumbnail\":false,\"namePrefix\":\"\",\"nameSuffix\":\"\"}},{\"fragments\":[{\"type\":\"message\",\"text\":\"foobarbaz\"}],\"attributes\":{\"firstName\":\"steve\",\"lastName\":\"stevenson\",\"hasPhoto\":true,\"numFields\":10,\"timesContacted\":5,\"lastTimeContacted\":0,\"isStarred\":false,\"hasCustomRingtone\":false,\"isSendToVoicemail\":false,\"hasThumbnail\":false,\"namePrefix\":\"\",\"nameSuffix\":\"\"}},{\"fragments\":[],\"attributes\":{\"firstName\":\"steve\",\"lastName\":\"stevenson\",\"hasPhoto\":true,\"numFields\":10,\"timesContacted\":5,\"lastTimeContacted\":0,\"isStarred\":false,\"hasCustomRingtone\":false,\"isSendToVoicemail\":false,\"hasThumbnail\":false,\"namePrefix\":\"\",\"nameSuffix\":\"\"}},{\"fragments\":[],\"attributes\":{\"firstName\":\"steve\",\"lastName\":\"stevenson\",\"hasPhoto\":true,\"numFields\":10,\"timesContacted\":5,\"lastTimeContacted\":0,\"isStarred\":false,\"hasCustomRingtone\":false,\"isSendToVoicemail\":false,\"hasThumbnail\":false,\"namePrefix\":\"\",\"nameSuffix\":\"\"}}],\"appType\":\"MY_APP\"}}")

func BenchmarkSaveContacts(b *testing.B) {
	gateway, err := benchGateway.CreateGateway(
		map[string]interface{}{
			"clients.baz.serviceName": "baz",
		},
		&testGateway.Options{
			KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
			KnownTChannelBackends: []string{"baz"},
			ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		},
		exampleGateway.CreateGateway,
	)
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
				b.Error("could not read response: " + res.Status)
				break
			}
			_ = res.Body.Close()
		}
	})

	// Test http method circuit breaker settings
	settings := hystrix.GetCircuitSettings()
	names := [2]string{"contacts-SaveContacts", "contacts-TestURLURL"}
	// contacts config values from production.json
	timeout := 10000
	max := 1000
	errorPercentage := 20
	// default parameters
	sleepWindow := 5000
	reqThreshold := 20
	expectedSettings := &hystrix.Settings{
		Timeout:                time.Duration(timeout) * time.Millisecond,
		MaxConcurrentRequests:  max,
		RequestVolumeThreshold: uint64(reqThreshold),
		SleepWindow:            time.Duration(sleepWindow) * time.Millisecond,
		ErrorPercentThreshold:  errorPercentage,
	}
	for _, name := range names {
		assert.Equal(b, settings[name], expectedSettings)
	}

	b.StopTimer()
	gateway.Close()
	b.StartTimer()
}

func TestSaveContactsCall(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"contacts"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
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

	saveContacts := &endpointContacts.Contacts_SaveContacts_Args{
		SaveContactsRequest: &endpointContacts.SaveContactsRequest{
			Contacts: []*endpointContacts.Contact{},
		},
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
