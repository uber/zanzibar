// Copyright (c) 2023 Uber Technologies, Inc.
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

package encoder

import (
	"strconv"
	"strings"
	"time"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Tags consists of list of Tag
type Tags []Tag

// Tag is a key value pair
type Tag struct {
	Key   string `json:"key,required"`
	Value string `json:"value,required"`
}

// StringTagEncoder is an ObjectEncoder backed by a slice []*Tag
type StringTagEncoder struct {
	fields   map[string]*Tag
	prefixes []string
}

// NewStringTagEncoder creates a new slice backed ObjectEncoder.
func NewStringTagEncoder(prefixes ...string) *StringTagEncoder {
	return &StringTagEncoder{
		fields:   map[string]*Tag{},
		prefixes: prefixes,
	}
}

// GetTags returns a slice of tags
func (s *StringTagEncoder) GetTags() []*Tag {
	tags := make([]*Tag, 0, len(s.fields))
	for _, t := range s.fields {
		tag := t
		tags = append(tags, tag)
	}
	return tags
}

// AddArray implements ObjectEncoder.
func (s *StringTagEncoder) AddArray(key string, v zapcore.ArrayMarshaler) error {
	arr := &sliceArrayEncoder{}
	if err := v.MarshalLogArray(arr); err != nil {
		return err
	}
	for _, elem := range arr.elems {
		field := zap.Any(key, elem)
		field.AddTo(s)
	}
	return nil
}

// AddObject implements ObjectEncoder.
func (s *StringTagEncoder) AddObject(k string, v zapcore.ObjectMarshaler) error {
	newMap := NewStringTagEncoder(k)
	if err := v.MarshalLogObject(newMap); err != nil {
		return err
	}
	for key, tag := range newMap.fields {
		s.fields[key] = tag
	}
	return nil
}

// AddBinary implements ObjectEncoder.
func (s *StringTagEncoder) AddBinary(k string, v []byte) { return } //encoder will ignore non string binaries

// AddByteString implements ObjectEncoder.
func (s *StringTagEncoder) AddByteString(k string, v []byte) { s.AddString(k, string(v)) }

// AddBool implements ObjectEncoder.
func (s *StringTagEncoder) AddBool(k string, v bool) { s.AddString(k, strconv.FormatBool(v)) }

// AddDuration implements ObjectEncoder.
func (s *StringTagEncoder) AddDuration(k string, v time.Duration) { s.AddString(k, v.String()) }

// AddComplex128 implements ObjectEncoder.
func (s *StringTagEncoder) AddComplex128(k string, v complex128) { return } //encoder will ignore complex128

// AddComplex64 implements ObjectEncoder.
func (s *StringTagEncoder) AddComplex64(k string, v complex64) { return } // encoder will ignore complex64

// AddFloat64 implements ObjectEncoder.
func (s *StringTagEncoder) AddFloat64(k string, v float64) {
	s.AddString(k, strconv.FormatFloat(v, 'f', -1, 64))
}

// AddFloat32 implements ObjectEncoder.
func (s *StringTagEncoder) AddFloat32(k string, v float32) { s.AddFloat64(k, float64(v)) }

// AddInt implements ObjectEncoder.
func (s *StringTagEncoder) AddInt(k string, v int) { s.AddString(k, strconv.Itoa(v)) }

// AddInt64 implements ObjectEncoder.
func (s *StringTagEncoder) AddInt64(k string, v int64) { s.AddInt(k, int(v)) }

// AddInt32 implements ObjectEncoder.
func (s *StringTagEncoder) AddInt32(k string, v int32) { s.AddInt(k, int(v)) }

// AddInt16 implements ObjectEncoder.
func (s *StringTagEncoder) AddInt16(k string, v int16) { s.AddInt(k, int(v)) }

// AddInt8 implements ObjectEncoder.
func (s *StringTagEncoder) AddInt8(k string, v int8) { s.AddInt(k, int(v)) }

// AddString implements ObjectEncoder.
func (s *StringTagEncoder) AddString(k string, v string) {
	s.fields[k] = &Tag{Key: s.key(k), Value: v}
}

// AddTime implements ObjectEncoder.
func (s *StringTagEncoder) AddTime(k string, v time.Time) { s.AddString(k, v.String()) }

// AddUint implements ObjectEncoder.
func (s *StringTagEncoder) AddUint(k string, v uint) { s.AddUint64(k, uint64(v)) }

// AddUint64 implements ObjectEncoder.
func (s *StringTagEncoder) AddUint64(k string, v uint64) { s.AddString(k, strconv.FormatUint(v, 10)) }

// AddUint32 implements ObjectEncoder.
func (s *StringTagEncoder) AddUint32(k string, v uint32) { s.AddUint64(k, uint64(v)) }

// AddUint16 implements ObjectEncoder.
func (s *StringTagEncoder) AddUint16(k string, v uint16) { s.AddUint64(k, uint64(v)) }

// AddUint8 implements ObjectEncoder.
func (s *StringTagEncoder) AddUint8(k string, v uint8) { s.AddUint64(k, uint64(v)) }

// AddUintptr implements ObjectEncoder.
func (s *StringTagEncoder) AddUintptr(k string, v uintptr) { return } //encoder will ignore uintptr

// AddReflected implements ObjectEncoder.
func (s *StringTagEncoder) AddReflected(k string, v interface{}) error { return nil } //encoder will ignore reflected values

// OpenNamespace implements ObjectEncoder.
func (s *StringTagEncoder) OpenNamespace(k string) {
	s.prefixes = append(s.prefixes, k)
}

func (s *StringTagEncoder) key(k string) string {
	return strings.Join(append(s.prefixes, k), ".")
}

// sliceArrayEncoder is an ArrayEncoder backed by a simple []interface{}. Like
// the MapObjectEncoder, it's not designed for production use.
type sliceArrayEncoder struct {
	elems []interface{}
}

func (s *sliceArrayEncoder) AppendArray(v zapcore.ArrayMarshaler) error {
	enc := &sliceArrayEncoder{}
	err := v.MarshalLogArray(enc)
	s.elems = append(s.elems, enc.elems)
	return err
}

func (s *sliceArrayEncoder) AppendObject(v zapcore.ObjectMarshaler) error {
	enc := NewStringTagEncoder()
	err := v.MarshalLogObject(enc)
	for _, f := range enc.fields {
		s.elems = append(s.elems, f.Value)
	}
	return err
}

func (s *sliceArrayEncoder) AppendReflected(v interface{}) error {
	s.elems = append(s.elems, v)
	return nil
}

func (s *sliceArrayEncoder) AppendBool(v bool)              { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendByteString(v []byte)      { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendComplex128(v complex128)  { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendComplex64(v complex64)    { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendDuration(v time.Duration) { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendFloat64(v float64)        { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendFloat32(v float32)        { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt(v int)                { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt64(v int64)            { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt32(v int32)            { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt16(v int16)            { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt8(v int8)              { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendString(v string)          { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendTime(v time.Time)         { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint(v uint)              { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint64(v uint64)          { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint32(v uint32)          { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint16(v uint16)          { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint8(v uint8)            { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUintptr(v uintptr)        { s.elems = append(s.elems, v) }
