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

package tchannel_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/uber/tchannel-go/testutils/testreader"
	"github.com/uber/tchannel-go/testutils/testwriter"
	"github.com/uber/zanzibar/runtime/tchannel"
	"github.com/uber/zanzibar/test/runtime/tchannel/gen-code/baz"
	"go.uber.org/thriftrw/protocol/binary"
	"go.uber.org/thriftrw/wire"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var structTest = struct {
	s       tchannel.RWTStruct
	encoded []byte
}{
	s: &baz.Data{
		B1: true,
		S2: "S2",
		I3: 3,
	},
	encoded: []byte{
		0x2,      // bool
		0x0, 0x1, // field 1
		0x1,      // true
		0xb,      // string
		0x0, 0x2, // field 2
		0x0, 0x0, 0x0, 0x2, // length of string "S2"
		'S', '2', // string "S2"
		0x8,      // i32
		0x0, 0x3, // field 3
		0x0, 0x0, 0x0, 0x3, // i32 3
		0x0, // end of struct
	},
}

type rwstructTest struct{}

func (rwstructTest) ToWire() (wire.Value, error) {
	return wire.Value{}, errors.New("ToWire error")
}

func (rwstructTest) FromWire(wire.Value) error {
	return errors.New("FromWire error")
}

func TestReadStruct(t *testing.T) {
	appendBytes := func(bs []byte, append []byte) []byte {
		b := make([]byte, len(bs)+len(append))
		n := copy(b, bs)
		copy(b[n:], append)
		return b
	}

	tests := []struct {
		s        tchannel.RWTStruct
		encoded  []byte
		wantErr  bool
		leftover []byte
	}{
		{
			s:       structTest.s,
			encoded: structTest.encoded,
		},
		{
			s: &baz.Data{
				B1: true,
				S2: "S2",
			},
			// Missing field 3, append byte 0 for end of struct.
			encoded: appendBytes(structTest.encoded[:len(structTest.encoded)-8], []byte{0}),
			wantErr: true,
		},
		{
			s:       structTest.s,
			encoded: appendBytes(structTest.encoded, []byte{1, 2, 3, 4}),
		},
	}

	for _, tt := range tests {
		s := &baz.Data{}
		readerAt := bytes.NewReader(tt.encoded)
		err := tchannel.ReadStruct(readerAt, s)

		assert.Equal(t, tt.wantErr, err != nil, "Unexpected error: %v", err)
		assert.Equal(t, tt.s, s, "Unexpected struct")
		assert.Equal(t, readerAt.Len(), len(tt.encoded), "readerAt data is not consumed")

		s = &baz.Data{}
		reader := bytes.NewBuffer(tt.encoded)
		err = tchannel.ReadStruct(reader, s)

		assert.Equal(t, tt.wantErr, err != nil, "Unexpected error: %v", err)
		assert.Equal(t, tt.s, s, "Unexpected struct")
		assert.Equal(t, 0, reader.Len(), "reader data is consumed")
	}
}

func TestReadStructErr(t *testing.T) {
	writer, reader := testreader.ChunkReader()
	writer <- structTest.encoded[:10]
	writer <- nil
	close(writer)

	s := &baz.Data{}
	err := tchannel.ReadStruct(reader, s)
	if assert.Error(t, err, "ReadStruct should fail") {
		// Apache Thrift just prepends the error message, and doesn't give us access
		// to the underlying error, so we can't check the underlying error exactly.
		assert.Contains(t, err.Error(), testreader.ErrUser.Error(), "Underlying error missing")
	}
}

func TestReadStructDecodeErr(t *testing.T) {
	reader := bytes.NewReader([]byte{1, 2, 3})
	s := &baz.Data{}
	err := tchannel.ReadStruct(reader, s)
	assert.Error(t, err, "ReadStruct should fail")
	assert.True(t, binary.IsDecodeError(err), "Should be decode error")
}

func TestWriteStruct(t *testing.T) {
	tests := []struct {
		s       tchannel.RWTStruct
		encoded []byte
		wantErr bool
	}{
		{
			s:       structTest.s,
			encoded: structTest.encoded,
		},
	}

	for _, tt := range tests {
		buf := &bytes.Buffer{}
		err := tchannel.WriteStruct(buf, tt.s)
		assert.Equal(t, tt.wantErr, err != nil, "Unexpected err: %v", err)
		if err != nil {
			continue
		}

		assert.Equal(t, tt.encoded, buf.Bytes(), "Encoded data mismatch")
	}
}

func TestWriteStructToWireErr(t *testing.T) {
	value := rwstructTest{}
	writer := testwriter.Limited(10)
	err := tchannel.WriteStruct(writer, value)
	assert.Error(t, err, "WriteStruct should fail")
	assert.Contains(t, err.Error(), "ToWire error", "Underlying error missing")
}

func TestWriteStructErr(t *testing.T) {
	writer := testwriter.Limited(10)
	err := tchannel.WriteStruct(writer, structTest.s)
	if assert.Error(t, err, "WriteStruct should fail") {
		// Apache Thrift just prepends the error message, and doesn't give us access
		// to the underlying error, so we can't check the underlying error exactly.
		assert.Contains(t, err.Error(), testwriter.ErrOutOfSpace.Error(), "Underlying error missing")
	}
}

func TestParallelReadWrites(t *testing.T) {
	var wg sync.WaitGroup
	testBG := func(f func(t *testing.T)) {
		wg.Add(1)
		go func() {
			f(t)
			wg.Done()
		}()
	}
	for i := 0; i < 50; i++ {
		testBG(TestReadStruct)
		testBG(TestWriteStruct)
	}
	wg.Wait()
}

func BenchmarkWriteStruct(b *testing.B) {
	buf := &bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		buf.Reset()
		err := tchannel.WriteStruct(buf, structTest.s)
		assert.NoError(b, err)
	}
}

func BenchmarkReadStruct(b *testing.B) {
	buf := bytes.NewReader(structTest.encoded)
	var d baz.Data

	_, err := buf.Seek(0, 0)
	assert.NoError(b, err)
	assert.NoError(b, tchannel.ReadStruct(buf, &d))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := buf.Seek(0, 0)
		assert.NoError(b, err)
		err = tchannel.ReadStruct(buf, &d)
		assert.NoError(b, err)
	}
}
