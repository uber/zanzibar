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
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-client-go"
	jzap "github.com/uber/jaeger-client-go/log/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestStringTagEncoder(t *testing.T) {
	timeString := time.Unix(0, 0)
	span := &jaeger.Span{}
	ctx := jaeger.NewSpanContext(
		jaeger.TraceID{High: 5678, Low: 1234},
		jaeger.SpanID(9),
		jaeger.SpanID(10),
		false,
		map[string]string{},
	)
	ptr := reflect.ValueOf(span)
	val := reflect.Indirect(ptr)
	//set operationName
	f := val.FieldByName("context")
	f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	v := reflect.ValueOf(ctx)
	f.Set(v)
	ctxWitSpan := opentracing.ContextWithSpan(context.Background(), span)
	tests := []struct {
		name     string
		prefixes []string
		fields   []zapcore.Field
		expected map[string][]string
	}{

		{
			name: "should encode string",
			fields: []zapcore.Field{
				zap.Any("foo", "bar"),
			},
			expected: map[string][]string{
				"foo": {"bar"},
			},
		},
		{
			name: "should encode array",
			fields: []zapcore.Field{
				zap.Strings("foo", []string{"bar", "boo"}),
			},
			expected: map[string][]string{
				"foo": {"bar", "boo"},
			},
		},
		{
			name: "should encode object",
			fields: []zapcore.Field{
				jzap.Trace(ctxWitSpan),
			},
			expected: map[string][]string{
				"trace.trace": {jaeger.TraceID{High: 5678, Low: 1234}.String()},
				"trace.span":  {jaeger.SpanID(9).String()},
			},
		},
		{
			name: "should encode with open namespace",
			fields: []zapcore.Field{
				zap.Any("foo", int64(1234)),
				{
					Key:  "boo",
					Type: zapcore.NamespaceType,
				},
				zap.Any("foo", int64(1234)),
			},
			expected: map[string][]string{
				"foo":     {"1234"},
				"boo.foo": {"1234"},
			},
		},
		{
			name: "should skip binary",
			fields: []zapcore.Field{
				zap.Any("foo", []uint8{}),
			},
			expected: map[string][]string{},
		},
		{
			name: "should encode binary string",
			fields: []zapcore.Field{
				{
					Key:       "foo",
					Type:      zapcore.ByteStringType,
					Interface: []byte("bar"),
				},
			},
			expected: map[string][]string{
				"foo": {"bar"},
			},
		},
		{
			name: "should encode bool",
			fields: []zapcore.Field{
				zap.Any("foo", true),
			},
			expected: map[string][]string{
				"foo": {"true"},
			},
		},
		{
			name: "should encode duration",
			fields: []zapcore.Field{
				zap.Any("foo", time.Nanosecond),
			},
			expected: map[string][]string{
				"foo": {"1ns"},
			},
		},
		{
			name: "should append errors",
			fields: []zapcore.Field{
				zap.Errors("errors", []error{errors.New("bar"), errors.New("boo")}),
			},
			expected: map[string][]string{
				"errors": {"bar", "boo"},
			},
		},
		{
			name: "should encode complex128",
			fields: []zapcore.Field{
				zap.Any("foo", complex128(12)),
			},
			expected: map[string][]string{},
		},
		{
			name: "should encode complex64",
			fields: []zapcore.Field{
				zap.Any("foo", complex64(12)),
			},
			expected: map[string][]string{},
		},
		{
			name: "should encode float64",
			fields: []zapcore.Field{
				zap.Any("foo", float64(12)),
			},
			expected: map[string][]string{
				"foo": {"12"},
			},
		},
		{
			name: "should encode float64",
			fields: []zapcore.Field{
				zap.Any("foo", float32(12)),
			},
			expected: map[string][]string{
				"foo": {"12"},
			},
		},
		{
			name: "should encode int",
			fields: []zapcore.Field{
				zap.Any("foo", int(1234)),
			},
			expected: map[string][]string{
				"foo": {"1234"},
			},
		},
		{
			name: "should encode int64",
			fields: []zapcore.Field{
				zap.Any("foo", int64(1234)),
			},
			expected: map[string][]string{
				"foo": {"1234"},
			},
		},
		{
			name: "should encode int32",
			fields: []zapcore.Field{
				zap.Any("foo", int32(1234)),
			},
			expected: map[string][]string{
				"foo": {"1234"},
			},
		},
		{
			name: "should encode int16",
			fields: []zapcore.Field{
				zap.Any("foo", int16(1)),
			},
			expected: map[string][]string{
				"foo": {"1"},
			},
		},
		{
			name: "should encode int8",
			fields: []zapcore.Field{
				zap.Any("foo", int8(1)),
			},
			expected: map[string][]string{
				"foo": {"1"},
			},
		},
		{
			name: "should encode uint",
			fields: []zapcore.Field{
				zap.Any("foo", uint(1)),
			},
			expected: map[string][]string{
				"foo": {"1"},
			},
		},
		{
			name: "should encode uint8",
			fields: []zapcore.Field{
				zap.Any("foo", uint8(1)),
			},
			expected: map[string][]string{
				"foo": {"1"},
			},
		},
		{
			name: "should encode uint16",
			fields: []zapcore.Field{
				zap.Any("foo", uint16(1)),
			},
			expected: map[string][]string{
				"foo": {"1"},
			},
		},
		{
			name: "should encode uint32",
			fields: []zapcore.Field{
				zap.Any("foo", uint32(1)),
			},
			expected: map[string][]string{
				"foo": {"1"},
			},
		},
		{
			name: "should encode uint64",
			fields: []zapcore.Field{
				zap.Any("foo", uint64(1)),
			},
			expected: map[string][]string{
				"foo": {"1"},
			},
		},
		{
			name: "should encode uintptr",
			fields: []zapcore.Field{
				zap.Any("foo", uintptr(1)),
			},
			expected: map[string][]string{
				"foo": {"1"},
			},
		},
		{
			name: "should encode time",
			fields: []zapcore.Field{
				zap.Any("foo", timeString),
			},
			expected: map[string][]string{
				"foo": {timeString.String()},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := NewStringTagEncoder(tt.prefixes...)
			for _, field := range tt.fields {
				field.AddTo(enc)
			}
			for _, field := range enc.fields {
				assert.Contains(t, tt.expected[field.Key], field.Value)
			}
		})
	}
}

func TestSliceArrayEncoder_AppendArray(t *testing.T) {
	enc := NewStringTagEncoder()
	senc := sliceArrayEncoder{}
	senc.AppendArray(zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
		inner.AppendBool(true)
		inner.AppendBool(false)
		return nil
	}))
	for _, elem := range senc.elems {
		field := zap.Any("foo", elem)
		field.AddTo(enc)
	}
	for _, f := range enc.fields {
		assert.Contains(t, []string{"true", "false"}, f.Value)
	}
}

func TestSliceArrayEncoder_Appends(t *testing.T) {
	enc := sliceArrayEncoder{}
	enc.AppendBool(true)
	enc.AppendByteString([]byte{})
	enc.AppendComplex64(complex64(1))
	enc.AppendComplex128(complex128(1))
	enc.AppendDuration(time.Second)
	enc.AppendFloat32(float32(1))
	enc.AppendFloat64(float64(1))
	enc.AppendInt(int(1))
	enc.AppendInt8(int8(1))
	enc.AppendInt16(int16(1))
	enc.AppendInt32(int32(1))
	enc.AppendInt64(int64(1))
	enc.AppendString("")
	enc.AppendTime(time.Now())
	enc.AppendUint(uint(1))
	enc.AppendUint8(uint8(1))
	enc.AppendUint16(uint16(1))
	enc.AppendUint32(uint32(1))
	enc.AppendUint64(uint64(1))
	enc.AppendUintptr(uintptr(1))
}
