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

package testm3server

import (
	"bytes"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	metrics "github.com/uber-go/tally/m3"
	customtransport "github.com/uber-go/tally/m3/customtransports"
	m3 "github.com/uber-go/tally/m3/thrift/v2"
	"github.com/uber-go/tally/thirdparty/github.com/apache/thrift/lib/go/thrift"
	"github.com/uber/zanzibar/test/lib"
)

var localListenAddr = &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)}

// FakeM3Server represents an M3 UDP server
type FakeM3Server struct {
	t         *testing.T
	Service   *FakeM3Service
	Addr      string
	protocol  metrics.Protocol
	processor thrift.TProcessor
	conn      *net.UDPConn
	closed    int32
}

// NewFakeM3Server creates server, pass it a wait group
// Which can be used to wait for receival of messages
func NewFakeM3Server(
	t *testing.T,
	wg *lib.WaitAtLeast,
	countBatches bool,
	countMetrics bool,
	protocol metrics.Protocol,
) *FakeM3Server {
	service := NewFakeM3Service(wg, countBatches, countMetrics)
	processor := m3.NewM3Processor(service)
	conn, err := net.ListenUDP(localListenAddr.Network(), localListenAddr)
	require.NoError(t, err, "ListenUDP failed")

	return &FakeM3Server{
		t:         t,
		Service:   service,
		Addr:      conn.LocalAddr().String(),
		conn:      conn,
		protocol:  protocol,
		processor: processor,
	}
}

// Serve will handle incoming UDP connections, currently only reads one packet
func (f *FakeM3Server) Serve() {
	readBuf := make([]byte, 64000)
	for f.conn != nil {
		n, err := f.conn.Read(readBuf)
		if err != nil {
			if atomic.LoadInt32(&f.closed) == 0 {
				f.t.Errorf("FakeM3Server failed to Read: %v", err)
			}
			return
		}

		trans, _ := customtransport.NewTBufferedReadTransport(bytes.NewBuffer(readBuf[0:n]))
		var proto thrift.TProtocol
		if f.protocol == metrics.Compact {
			proto = thrift.NewTCompactProtocol(trans)
		} else {
			proto = thrift.NewTBinaryProtocolTransport(trans)
		}
		_, _ = f.processor.Process(proto, proto)
	}
}

// Close closes the UDP server
func (f *FakeM3Server) Close() error {
	atomic.AddInt32(&f.closed, 1)
	return f.conn.Close()
}

// NewFakeM3Service creates an M3Service
func NewFakeM3Service(
	wg *lib.WaitAtLeast,
	countBatches bool,
	countMetrics bool,
) *FakeM3Service {
	return &FakeM3Service{
		metrics:      make(map[string]m3.Metric),
		wg:           wg,
		countBatches: countBatches,
		countMetrics: countMetrics,
	}
}

// FakeM3Service tracks the metrics the server has received
type FakeM3Service struct {
	MaxMetrics       int
	seenMetricsCount int
	lock             sync.RWMutex
	batches          []m3.MetricBatch
	metrics          map[string]m3.Metric
	wg               *lib.WaitAtLeast
	countBatches     bool
	countMetrics     bool
}

// GetBatches gets the batches
func (m *FakeM3Service) GetBatches() []m3.MetricBatch {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.batches
}

// GetMetrics gets the individual metrics
func (m *FakeM3Service) GetMetrics() map[string]m3.Metric {
	// sleep to avoid race conditions with m3.Client
	// Basically m3.Client can still be emitting and we might not
	// yet have all metrics in `m.metrics`
	time.Sleep(100 * time.Millisecond)

	m.lock.RLock()
	defer m.lock.RUnlock()

	cloneMap := map[string]m3.Metric{}
	for key, metric := range m.metrics {
		cloneMap[key] = metric
	}

	return cloneMap
}

// EmitMetricBatchV2 is called by thrift message processor
func (m *FakeM3Service) EmitMetricBatchV2(batch m3.MetricBatch) (err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.batches = append(m.batches, batch)
	if m.countBatches {
		m.wg.Done()
	}

	for _, metric := range batch.Metrics {
		if m.countMetrics && m.MaxMetrics != 0 {
			if m.seenMetricsCount == m.MaxMetrics {
				continue
			}

			seen := m.storeMetric(metric)
			if !seen {
				m.seenMetricsCount++
			}
			if !seen && m.countMetrics {
				m.wg.Done()
			}
		} else {
			m.storeMetric(metric)
			if m.countMetrics {
				m.wg.Done()
			}
		}

	}

	return thrift.NewTTransportException(thrift.END_OF_FILE, "complete")
}

func (m *FakeM3Service) storeMetric(metric m3.Metric) bool {
	var (
		mTags = metric.GetTags()
		tags  = make(map[string]string, len(mTags))
	)
	for _, tag := range mTags {
		tags[tag.Name] = tag.Value
	}

	var (
		key     = tally.KeyForPrefixedStringMap(metric.Name, tags)
		_, seen = m.metrics[key]
	)

	metric.Value.Count += m.metrics[key].Value.Count
	m.metrics[key] = metric

	return seen
}
