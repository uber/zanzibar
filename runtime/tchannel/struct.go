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

package tchannel

import (
	"bytes"
	"io"

	"code.uber.internal/rt/tchan-server/thriftrw/buffer"
	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"
)

// WriteStruct writes the given Thriftrw struct to a writer.
func WriteStruct(writer io.Writer, s RWTStruct) error {
	wireValue, err := s.ToWire()
	if err != nil {
		return err
	}

	if err := protocol.Binary.Encode(wireValue, writer); err != nil {
		return err
	}

	return nil
}

// ReadStruct reads the given Thriftrw struct.
func ReadStruct(reader io.Reader, s RWTStruct) error {
	readerAt, ok := reader.(io.ReaderAt)

	// do not read all to buffer if reader already is type of io.ReaderAt
	if !ok {
		buf := buffer.Get()
		defer buffer.Put(buf)

		if _, err := buf.ReadFrom(reader); err != nil {
			return err
		}
		readerAt = bytes.NewReader(buf.Bytes())
	}

	wireValue, err := protocol.Binary.Decode(readerAt, wire.TStruct)
	if err != nil {
		return err
	}
	err = s.FromWire(wireValue)
	return err
}
