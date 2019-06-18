// Code generated by thriftrw v1.19.0. DO NOT EDIT.
// @generated

package baz

import (
	fmt "fmt"
	wire "go.uber.org/thriftrw/wire"
	zapcore "go.uber.org/zap/zapcore"
	strings "strings"
)

// SimpleService_TestUuid_Args represents the arguments for the SimpleService.testUuid function.
//
// The arguments for testUuid are sent and received over the wire as this struct.
type SimpleService_TestUuid_Args struct {
}

// ToWire translates a SimpleService_TestUuid_Args struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//   x, err := v.ToWire()
//   if err != nil {
//     return err
//   }
//
//   if err := binaryProtocol.Encode(x, writer); err != nil {
//     return err
//   }
func (v *SimpleService_TestUuid_Args) ToWire() (wire.Value, error) {
	var (
		fields [0]wire.Field
		i      int = 0
	)

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a SimpleService_TestUuid_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a SimpleService_TestUuid_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v SimpleService_TestUuid_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *SimpleService_TestUuid_Args) FromWire(w wire.Value) error {

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		}
	}

	return nil
}

// String returns a readable string representation of a SimpleService_TestUuid_Args
// struct.
func (v *SimpleService_TestUuid_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [0]string
	i := 0

	return fmt.Sprintf("SimpleService_TestUuid_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this SimpleService_TestUuid_Args match the
// provided SimpleService_TestUuid_Args.
//
// This function performs a deep comparison.
func (v *SimpleService_TestUuid_Args) Equals(rhs *SimpleService_TestUuid_Args) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of SimpleService_TestUuid_Args.
func (v *SimpleService_TestUuid_Args) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	return err
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "testUuid" for this struct.
func (v *SimpleService_TestUuid_Args) MethodName() string {
	return "testUuid"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *SimpleService_TestUuid_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// SimpleService_TestUuid_Helper provides functions that aid in handling the
// parameters and return values of the SimpleService.testUuid
// function.
var SimpleService_TestUuid_Helper = struct {
	// Args accepts the parameters of testUuid in-order and returns
	// the arguments struct for the function.
	Args func() *SimpleService_TestUuid_Args

	// IsException returns true if the given error can be thrown
	// by testUuid.
	//
	// An error can be thrown by testUuid only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for testUuid
	// given the error returned by it. The provided error may
	// be nil if testUuid did not fail.
	//
	// This allows mapping errors returned by testUuid into a
	// serializable result struct. WrapResponse returns a
	// non-nil error if the provided error cannot be thrown by
	// testUuid
	//
	//   err := testUuid(args)
	//   result, err := SimpleService_TestUuid_Helper.WrapResponse(err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from testUuid: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(error) (*SimpleService_TestUuid_Result, error)

	// UnwrapResponse takes the result struct for testUuid
	// and returns the erorr returned by it (if any).
	//
	// The error is non-nil only if testUuid threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   err := SimpleService_TestUuid_Helper.UnwrapResponse(result)
	UnwrapResponse func(*SimpleService_TestUuid_Result) error
}{}

func init() {
	SimpleService_TestUuid_Helper.Args = func() *SimpleService_TestUuid_Args {
		return &SimpleService_TestUuid_Args{}
	}

	SimpleService_TestUuid_Helper.IsException = func(err error) bool {
		switch err.(type) {
		default:
			return false
		}
	}

	SimpleService_TestUuid_Helper.WrapResponse = func(err error) (*SimpleService_TestUuid_Result, error) {
		if err == nil {
			return &SimpleService_TestUuid_Result{}, nil
		}

		return nil, err
	}
	SimpleService_TestUuid_Helper.UnwrapResponse = func(result *SimpleService_TestUuid_Result) (err error) {
		return
	}

}

// SimpleService_TestUuid_Result represents the result of a SimpleService.testUuid function call.
//
// The result of a testUuid execution is sent and received over the wire as this struct.
type SimpleService_TestUuid_Result struct {
}

// ToWire translates a SimpleService_TestUuid_Result struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//   x, err := v.ToWire()
//   if err != nil {
//     return err
//   }
//
//   if err := binaryProtocol.Encode(x, writer); err != nil {
//     return err
//   }
func (v *SimpleService_TestUuid_Result) ToWire() (wire.Value, error) {
	var (
		fields [0]wire.Field
		i      int = 0
	)

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a SimpleService_TestUuid_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a SimpleService_TestUuid_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v SimpleService_TestUuid_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *SimpleService_TestUuid_Result) FromWire(w wire.Value) error {

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		}
	}

	return nil
}

// String returns a readable string representation of a SimpleService_TestUuid_Result
// struct.
func (v *SimpleService_TestUuid_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [0]string
	i := 0

	return fmt.Sprintf("SimpleService_TestUuid_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this SimpleService_TestUuid_Result match the
// provided SimpleService_TestUuid_Result.
//
// This function performs a deep comparison.
func (v *SimpleService_TestUuid_Result) Equals(rhs *SimpleService_TestUuid_Result) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of SimpleService_TestUuid_Result.
func (v *SimpleService_TestUuid_Result) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	return err
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "testUuid" for this struct.
func (v *SimpleService_TestUuid_Result) MethodName() string {
	return "testUuid"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *SimpleService_TestUuid_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
