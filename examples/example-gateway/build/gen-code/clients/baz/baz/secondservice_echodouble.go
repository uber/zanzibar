// Code generated by thriftrw v1.9.0. DO NOT EDIT.
// @generated

package baz

import (
	"errors"
	"fmt"
	"go.uber.org/thriftrw/wire"
	"strings"
)

// SecondService_EchoDouble_Args represents the arguments for the SecondService.echoDouble function.
//
// The arguments for echoDouble are sent and received over the wire as this struct.
type SecondService_EchoDouble_Args struct {
	Arg float64 `json:"arg,required"`
}

// ToWire translates a SecondService_EchoDouble_Args struct into a Thrift-level intermediate
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
func (v *SecondService_EchoDouble_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueDouble(v.Arg), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a SecondService_EchoDouble_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a SecondService_EchoDouble_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v SecondService_EchoDouble_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *SecondService_EchoDouble_Args) FromWire(w wire.Value) error {
	var err error

	argIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TDouble {
				v.Arg, err = field.Value.GetDouble(), error(nil)
				if err != nil {
					return err
				}
				argIsSet = true
			}
		}
	}

	if !argIsSet {
		return errors.New("field Arg of SecondService_EchoDouble_Args is required")
	}

	return nil
}

// String returns a readable string representation of a SecondService_EchoDouble_Args
// struct.
func (v *SecondService_EchoDouble_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	fields[i] = fmt.Sprintf("Arg: %v", v.Arg)
	i++

	return fmt.Sprintf("SecondService_EchoDouble_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this SecondService_EchoDouble_Args match the
// provided SecondService_EchoDouble_Args.
//
// This function performs a deep comparison.
func (v *SecondService_EchoDouble_Args) Equals(rhs *SecondService_EchoDouble_Args) bool {
	if !(v.Arg == rhs.Arg) {
		return false
	}

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "echoDouble" for this struct.
func (v *SecondService_EchoDouble_Args) MethodName() string {
	return "echoDouble"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *SecondService_EchoDouble_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// SecondService_EchoDouble_Helper provides functions that aid in handling the
// parameters and return values of the SecondService.echoDouble
// function.
var SecondService_EchoDouble_Helper = struct {
	// Args accepts the parameters of echoDouble in-order and returns
	// the arguments struct for the function.
	Args func(
		arg float64,
	) *SecondService_EchoDouble_Args

	// IsException returns true if the given error can be thrown
	// by echoDouble.
	//
	// An error can be thrown by echoDouble only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for echoDouble
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// echoDouble into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by echoDouble
	//
	//   value, err := echoDouble(args)
	//   result, err := SecondService_EchoDouble_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from echoDouble: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(float64, error) (*SecondService_EchoDouble_Result, error)

	// UnwrapResponse takes the result struct for echoDouble
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if echoDouble threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := SecondService_EchoDouble_Helper.UnwrapResponse(result)
	UnwrapResponse func(*SecondService_EchoDouble_Result) (float64, error)
}{}

func init() {
	SecondService_EchoDouble_Helper.Args = func(
		arg float64,
	) *SecondService_EchoDouble_Args {
		return &SecondService_EchoDouble_Args{
			Arg: arg,
		}
	}

	SecondService_EchoDouble_Helper.IsException = func(err error) bool {
		switch err.(type) {
		default:
			return false
		}
	}

	SecondService_EchoDouble_Helper.WrapResponse = func(success float64, err error) (*SecondService_EchoDouble_Result, error) {
		if err == nil {
			return &SecondService_EchoDouble_Result{Success: &success}, nil
		}

		return nil, err
	}
	SecondService_EchoDouble_Helper.UnwrapResponse = func(result *SecondService_EchoDouble_Result) (success float64, err error) {

		if result.Success != nil {
			success = *result.Success
			return
		}

		err = errors.New("expected a non-void result")
		return
	}

}

// SecondService_EchoDouble_Result represents the result of a SecondService.echoDouble function call.
//
// The result of a echoDouble execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type SecondService_EchoDouble_Result struct {
	// Value returned by echoDouble after a successful execution.
	Success *float64 `json:"success,omitempty"`
}

// ToWire translates a SecondService_EchoDouble_Result struct into a Thrift-level intermediate
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
func (v *SecondService_EchoDouble_Result) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Success != nil {
		w, err = wire.NewValueDouble(*(v.Success)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 0, Value: w}
		i++
	}

	if i != 1 {
		return wire.Value{}, fmt.Errorf("SecondService_EchoDouble_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a SecondService_EchoDouble_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a SecondService_EchoDouble_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v SecondService_EchoDouble_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *SecondService_EchoDouble_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TDouble {
				var x float64
				x, err = field.Value.GetDouble(), error(nil)
				v.Success = &x
				if err != nil {
					return err
				}

			}
		}
	}

	count := 0
	if v.Success != nil {
		count++
	}
	if count != 1 {
		return fmt.Errorf("SecondService_EchoDouble_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a SecondService_EchoDouble_Result
// struct.
func (v *SecondService_EchoDouble_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.Success != nil {
		fields[i] = fmt.Sprintf("Success: %v", *(v.Success))
		i++
	}

	return fmt.Sprintf("SecondService_EchoDouble_Result{%v}", strings.Join(fields[:i], ", "))
}

func _Double_EqualsPtr(lhs, rhs *float64) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

// Equals returns true if all the fields of this SecondService_EchoDouble_Result match the
// provided SecondService_EchoDouble_Result.
//
// This function performs a deep comparison.
func (v *SecondService_EchoDouble_Result) Equals(rhs *SecondService_EchoDouble_Result) bool {
	if !_Double_EqualsPtr(v.Success, rhs.Success) {
		return false
	}

	return true
}

// GetSuccess returns the value of Success if it is set or its
// zero value if it is unset.
func (v *SecondService_EchoDouble_Result) GetSuccess() (o float64) {
	if v.Success != nil {
		return *v.Success
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "echoDouble" for this struct.
func (v *SecondService_EchoDouble_Result) MethodName() string {
	return "echoDouble"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *SecondService_EchoDouble_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
