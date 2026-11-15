// Copyright (c) 2024 Uber Technologies, Inc.
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

package zanzibar

import (
	"bytes"
	"io"
	"math/rand"
	"sync"
	"testing"

	"github.com/uber/tchannel-go/testutils/testreader"
	clientsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/bar/bar"
	"go.uber.org/thriftrw/protocol/binary"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureEmptySuccess(t *testing.T) {
	reader := bytes.NewReader(nil)
	err := EnsureEmpty(reader, "success")
	require.NoError(t, err, "ensureEmpty should succeed with empty reader")
}

func TestEnsureEmptyHasBytes(t *testing.T) {
	reader := bytes.NewReader([]byte{1, 2, 3})
	err := EnsureEmpty(reader, "T")
	require.Error(t, err, "ensureEmpty should fail when there's bytes")
	assert.Equal(t, err.Error(), "found unexpected bytes after T, found (upto 128 bytes): 010203")
}

func TestEnsureEmptyError(t *testing.T) {
	control, reader := testreader.ChunkReader()
	control <- nil
	close(control)

	err := EnsureEmpty(reader, "has bytes")
	require.Error(t, err, "ensureEmpty should fail when there's an error")
	assert.Equal(t, testreader.ErrUser, err, "Unexpected error")
}

func TestBuffers(t *testing.T) {
	var wg sync.WaitGroup
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 100; i++ {
				buf := GetBuffer()
				assert.Zero(t, buf.Len(), "Expected truncated buffer")

				bytesOfNoise := make([]byte, rand.Intn(5000))
				_, err := rand.Read(bytesOfNoise)
				assert.NoError(t, err, "Unexpected error from rand.Read")
				_, err = buf.Write(bytesOfNoise)
				assert.NoError(t, err, "Unexpected error from buffer.Write")

				assert.Equal(t, buf.Len(), len(bytesOfNoise), "Expected same buffer size")

				PutBuffer(buf)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

// fromWireOnly hides Decode, so ReadStruct takes the FromWire path.
type fromWireOnly struct{ RWTStruct }

func BenchmarkReadStruct(b *testing.B) {
	args := &clientsBarBar.Echo_EchoStructList_Args{}
	for i := 0; i < 8; i++ {
		args.Arg = append(args.Arg, &clientsBarBar.BarResponse{
			StringField:        "a",
			MapIntWithRange:    map[clientsBarBar.UUID]int32{"fakeUUID": int32(i)},
			MapIntWithoutRange: map[string]int32{"0": int32(i)},
			BinaryField:        []byte("d29ybGQ="),
		})
	}
	w, err := args.ToWire()
	require.NoError(b, err)
	var buf bytes.Buffer
	require.NoError(b, binary.Default.Encode(w, &buf))
	data := buf.Bytes()

	benchmarks := []struct {
		name      string
		newStruct func() RWTStruct
	}{
		{"Decode", func() RWTStruct { return &clientsBarBar.Echo_EchoStructList_Args{} }},
		{"FromWire", func() RWTStruct { return fromWireOnly{&clientsBarBar.Echo_EchoStructList_Args{}} }},
	}
	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				if err := ReadStruct(struct{ io.Reader }{bytes.NewReader(data)}, bb.newStruct()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
