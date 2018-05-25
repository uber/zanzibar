// Code generated by thriftrw v1.8.0. DO NOT EDIT.
// @generated

package baz

import (
	"errors"
	"fmt"
	"go.uber.org/thriftrw/wire"
	"strings"
)

// SimpleService_Call_Args represents the arguments for the SimpleService.call function.
//
// The arguments for call are sent and received over the wire as this struct.
type SimpleService_Call_Args struct {
	Arg         *BazRequest `json:"arg,required"`
	I64Optional *int64      `json:"i64Optional,omitempty"`
	TestUUID    *UUID       `json:"testUUID,omitempty"`
}

// ToWire translates a SimpleService_Call_Args struct into a Thrift-level intermediate
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
func (v *SimpleService_Call_Args) ToWire() (wire.Value, error) {
	var (
		fields [3]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Arg == nil {
		return w, errors.New("field Arg of SimpleService_Call_Args is required")
	}
	w, err = v.Arg.ToWire()
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++
	if v.I64Optional != nil {
		w, err = wire.NewValueI64(*(v.I64Optional)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 2, Value: w}
		i++
	}
	if v.TestUUID != nil {
		w, err = v.TestUUID.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 3, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _BazRequest_Read(w wire.Value) (*BazRequest, error) {
	var v BazRequest
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a SimpleService_Call_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a SimpleService_Call_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v SimpleService_Call_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *SimpleService_Call_Args) FromWire(w wire.Value) error {
	var err error

	argIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.Arg, err = _BazRequest_Read(field.Value)
				if err != nil {
					return err
				}
				argIsSet = true
			}
		case 2:
			if field.Value.Type() == wire.TI64 {
				var x int64
				x, err = field.Value.GetI64(), error(nil)
				v.I64Optional = &x
				if err != nil {
					return err
				}

			}
		case 3:
			if field.Value.Type() == wire.TBinary {
				var x UUID
				x, err = _UUID_Read(field.Value)
				v.TestUUID = &x
				if err != nil {
					return err
				}

			}
		}
	}

	if !argIsSet {
		return errors.New("field Arg of SimpleService_Call_Args is required")
	}

	return nil
}

// String returns a readable string representation of a SimpleService_Call_Args
// struct.
func (v *SimpleService_Call_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [3]string
	i := 0
	fields[i] = fmt.Sprintf("Arg: %v", v.Arg)
	i++
	if v.I64Optional != nil {
		fields[i] = fmt.Sprintf("I64Optional: %v", *(v.I64Optional))
		i++
	}
	if v.TestUUID != nil {
		fields[i] = fmt.Sprintf("TestUUID: %v", *(v.TestUUID))
		i++
	}

	return fmt.Sprintf("SimpleService_Call_Args{%v}", strings.Join(fields[:i], ", "))
}

func _I64_EqualsPtr(lhs, rhs *int64) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

func _UUID_EqualsPtr(lhs, rhs *UUID) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

// Equals returns true if all the fields of this SimpleService_Call_Args match the
// provided SimpleService_Call_Args.
//
// This function performs a deep comparison.
func (v *SimpleService_Call_Args) Equals(rhs *SimpleService_Call_Args) bool {
	if !v.Arg.Equals(rhs.Arg) {
		return false
	}
	if !_I64_EqualsPtr(v.I64Optional, rhs.I64Optional) {
		return false
	}
	if !_UUID_EqualsPtr(v.TestUUID, rhs.TestUUID) {
		return false
	}

	return true
}

// GetI64Optional returns the value of I64Optional if it is set or its
// zero value if it is unset.
func (v *SimpleService_Call_Args) GetI64Optional() (o int64) {
	if v.I64Optional != nil {
		return *v.I64Optional
	}

	return
}

// GetTestUUID returns the value of TestUUID if it is set or its
// zero value if it is unset.
func (v *SimpleService_Call_Args) GetTestUUID() (o UUID) {
	if v.TestUUID != nil {
		return *v.TestUUID
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "call" for this struct.
func (v *SimpleService_Call_Args) MethodName() string {
	return "call"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *SimpleService_Call_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// SimpleService_Call_Helper provides functions that aid in handling the
// parameters and return values of the SimpleService.call
// function.
var SimpleService_Call_Helper = struct {
	// Args accepts the parameters of call in-order and returns
	// the arguments struct for the function.
	Args func(
		arg *BazRequest,
		i64Optional *int64,
		testUUID *UUID,
	) *SimpleService_Call_Args

	// IsException returns true if the given error can be thrown
	// by call.
	//
	// An error can be thrown by call only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for call
	// given the error returned by it. The provided error may
	// be nil if call did not fail.
	//
	// This allows mapping errors returned by call into a
	// serializable result struct. WrapResponse returns a
	// non-nil error if the provided error cannot be thrown by
	// call
	//
	//   err := call(args)
	//   result, err := SimpleService_Call_Helper.WrapResponse(err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from call: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(error) (*SimpleService_Call_Result, error)

	// UnwrapResponse takes the result struct for call
	// and returns the erorr returned by it (if any).
	//
	// The error is non-nil only if call threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   err := SimpleService_Call_Helper.UnwrapResponse(result)
	UnwrapResponse func(*SimpleService_Call_Result) error
}{}

func init() {
	SimpleService_Call_Helper.Args = func(
		arg *BazRequest,
		i64Optional *int64,
		testUUID *UUID,
	) *SimpleService_Call_Args {
		return &SimpleService_Call_Args{
			Arg:         arg,
			I64Optional: i64Optional,
			TestUUID:    testUUID,
		}
	}

	SimpleService_Call_Helper.IsException = func(err error) bool {
		switch err.(type) {
		case *AuthErr:
			return true
		default:
			return false
		}
	}

	SimpleService_Call_Helper.WrapResponse = func(err error) (*SimpleService_Call_Result, error) {
		if err == nil {
			return &SimpleService_Call_Result{}, nil
		}

		switch e := err.(type) {
		case *AuthErr:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for SimpleService_Call_Result.AuthErr")
			}
			return &SimpleService_Call_Result{AuthErr: e}, nil
		}

		return nil, err
	}
	SimpleService_Call_Helper.UnwrapResponse = func(result *SimpleService_Call_Result) (err error) {
		if result.AuthErr != nil {
			err = result.AuthErr
			return
		}
		return
	}

}

// SimpleService_Call_Result represents the result of a SimpleService.call function call.
//
// The result of a call execution is sent and received over the wire as this struct.
type SimpleService_Call_Result struct {
	AuthErr *AuthErr `json:"authErr,omitempty"`
}

// ToWire translates a SimpleService_Call_Result struct into a Thrift-level intermediate
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
func (v *SimpleService_Call_Result) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.AuthErr != nil {
		w, err = v.AuthErr.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}

	if i > 1 {
		return wire.Value{}, fmt.Errorf("SimpleService_Call_Result should have at most one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _AuthErr_Read(w wire.Value) (*AuthErr, error) {
	var v AuthErr
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a SimpleService_Call_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a SimpleService_Call_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v SimpleService_Call_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *SimpleService_Call_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.AuthErr, err = _AuthErr_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	count := 0
	if v.AuthErr != nil {
		count++
	}
	if count > 1 {
		return fmt.Errorf("SimpleService_Call_Result should have at most one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a SimpleService_Call_Result
// struct.
func (v *SimpleService_Call_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.AuthErr != nil {
		fields[i] = fmt.Sprintf("AuthErr: %v", v.AuthErr)
		i++
	}

	return fmt.Sprintf("SimpleService_Call_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this SimpleService_Call_Result match the
// provided SimpleService_Call_Result.
//
// This function performs a deep comparison.
func (v *SimpleService_Call_Result) Equals(rhs *SimpleService_Call_Result) bool {
	if !((v.AuthErr == nil && rhs.AuthErr == nil) || (v.AuthErr != nil && rhs.AuthErr != nil && v.AuthErr.Equals(rhs.AuthErr))) {
		return false
	}

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "call" for this struct.
func (v *SimpleService_Call_Result) MethodName() string {
	return "call"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *SimpleService_Call_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
