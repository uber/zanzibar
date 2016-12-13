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
	"io/ioutil"
	"testing"

	"bytes"

	"net/http"

	assert "github.com/stretchr/testify/assert"
	config "github.com/uber/zanzibar/examples/example-gateway/config"
	contacts "github.com/uber/zanzibar/examples/example-gateway/endpoints/contacts"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testBackend "github.com/uber/zanzibar/test/lib/test_backend"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
)

var benchBytes = []byte("{\"contacts\":[{\"fragments\":[{\"type\":\"message\",\"text\":\"foobarbaz\"}],\"attributes\":{\"firstName\":\"steve\",\"lastName\":\"stevenson\",\"hasPhoto\":true,\"numFields\":10,\"timesContacted\":5,\"lastTimeContacted\":0,\"isStarred\":false,\"hasCustomRingtone\":false,\"isSendToVoicemail\":false,\"hasThumbnail\":false,\"namePrefix\":\"\",\"nameSuffix\":\"\"}},{\"fragments\":[{\"type\":\"message\",\"text\":\"foobarbaz\"}],\"attributes\":{\"firstName\":\"steve\",\"lastName\":\"stevenson\",\"hasPhoto\":true,\"numFields\":10,\"timesContacted\":5,\"lastTimeContacted\":0,\"isStarred\":false,\"hasCustomRingtone\":false,\"isSendToVoicemail\":false,\"hasThumbnail\":false,\"namePrefix\":\"\",\"nameSuffix\":\"\"}},{\"fragments\":[],\"attributes\":{\"firstName\":\"steve\",\"lastName\":\"stevenson\",\"hasPhoto\":true,\"numFields\":10,\"timesContacted\":5,\"lastTimeContacted\":0,\"isStarred\":false,\"hasCustomRingtone\":false,\"isSendToVoicemail\":false,\"hasThumbnail\":false,\"namePrefix\":\"\",\"nameSuffix\":\"\"}},{\"fragments\":[],\"attributes\":{\"firstName\":\"steve\",\"lastName\":\"stevenson\",\"hasPhoto\":true,\"numFields\":10,\"timesContacted\":5,\"lastTimeContacted\":0,\"isStarred\":false,\"hasCustomRingtone\":false,\"isSendToVoicemail\":false,\"hasThumbnail\":false,\"namePrefix\":\"\",\"nameSuffix\":\"\"}}],\"appType\":\"MY_APP\"}")

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

	handleContacts := func(w http.ResponseWriter, r *http.Request) {
		testCase.Counter++
		w.WriteHeader(202)
	}
	testCase.Backend.HandleFunc("POST", "/foo/contacts", handleContacts)

	config := &config.Config{}
	config.Clients.Contacts.IP = "127.0.0.1"
	config.Clients.Contacts.Port = testCase.Backend.RealPort

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

func BenchmarkSaveContacts(b *testing.B) {
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
				"POST", "/contacts/foo/contacts",
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
		}
	})

	b.StopTimer()
	testCase.Close()
	b.StartTimer()
}

func TestSaveContactsCall(t *testing.T) {
	testCase, err := newTestCase(t, false)
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer testCase.Close()

	assert.NotNil(t, testCase.TestGateway, "gateway exists")

	saveContacts := &contacts.SaveContactsRequest{
		Contacts: []*contacts.Contact{},
	}
	rawBody, _ := saveContacts.MarshalJSON()

	res, err := testCase.TestGateway.MakeRequest(
		"POST", "/contacts/foo/contacts", bytes.NewReader(rawBody),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "202 Accepted", res.Status)
	assert.Equal(t, 1, testCase.Counter)
}
