package foo_test

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

var benchBytes = []byte(<no value>})

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

    generated := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte({"statusCode":200})); err != nil {
			t.Fatal("can't write fake response")
		}
		testCase.Counter++
	}
	testCase.Backend.HandleFunc(POST, /add-credentials, generated)

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

func Benchmark<no value>Foo(b *testing.B) {
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
		POST, /googlenow/add-credentials, bytes.NewReader(benchBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, testCase.Counter)
}
