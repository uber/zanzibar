// Code generated by thriftrw v1.23.0. DO NOT EDIT.
// @generated

package quux

import (
	errors "errors"
	fmt "fmt"
	strings "strings"

	wire "go.uber.org/thriftrw/wire"
	zapcore "go.uber.org/zap/zapcore"
)

// SimpleService_EchoString_Args represents the arguments for the SimpleService.EchoString function.
//
// The arguments for EchoString are sent and received over the wire as this struct.
type SimpleService_EchoString_Args struct {
	Msg string `json:"msg,required"`
}

// ToWire translates a SimpleService_EchoString_Args struct into a Thrift-level intermediate
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
func (v *SimpleService_EchoString_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueString(v.Msg), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a SimpleService_EchoString_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a SimpleService_EchoString_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v SimpleService_EchoString_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *SimpleService_EchoString_Args) FromWire(w wire.Value) error {
	var err error

	msgIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.Msg, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				msgIsSet = true
			}
		}
	}

	if !msgIsSet {
		return errors.New("field Msg of SimpleService_EchoString_Args is required")
	}

	return nil
}

// String returns a readable string representation of a SimpleService_EchoString_Args
// struct.
func (v *SimpleService_EchoString_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	fields[i] = fmt.Sprintf("Msg: %v", v.Msg)
	i++

	return fmt.Sprintf("SimpleService_EchoString_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this SimpleService_EchoString_Args match the
// provided SimpleService_EchoString_Args.
//
// This function performs a deep comparison.
func (v *SimpleService_EchoString_Args) Equals(rhs *SimpleService_EchoString_Args) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !(v.Msg == rhs.Msg) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of SimpleService_EchoString_Args.
func (v *SimpleService_EchoString_Args) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	enc.AddString("msg", v.Msg)
	return err
}

// GetMsg returns the value of Msg if it is set or its
// zero value if it is unset.
func (v *SimpleService_EchoString_Args) GetMsg() (o string) {
	if v != nil {
		o = v.Msg
	}
	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "EchoString" for this struct.
func (v *SimpleService_EchoString_Args) MethodName() string {
	return "EchoString"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *SimpleService_EchoString_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// SimpleService_EchoString_Helper provides functions that aid in handling the
// parameters and return values of the SimpleService.EchoString
// function.
var SimpleService_EchoString_Helper = struct {
	// Args accepts the parameters of EchoString in-order and returns
	// the arguments struct for the function.
	Args func(
		msg string,
	) *SimpleService_EchoString_Args

	// IsException returns true if the given error can be thrown
	// by EchoString.
	//
	// An error can be thrown by EchoString only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for EchoString
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// EchoString into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by EchoString
	//
	//   value, err := EchoString(args)
	//   result, err := SimpleService_EchoString_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from EchoString: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(string, error) (*SimpleService_EchoString_Result, error)

	// UnwrapResponse takes the result struct for EchoString
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if EchoString threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := SimpleService_EchoString_Helper.UnwrapResponse(result)
	UnwrapResponse func(*SimpleService_EchoString_Result) (string, error)
}{}

func init() {
	SimpleService_EchoString_Helper.Args = func(
		msg string,
	) *SimpleService_EchoString_Args {
		return &SimpleService_EchoString_Args{
			Msg: msg,
		}
	}

	SimpleService_EchoString_Helper.IsException = func(err error) bool {
		switch err.(type) {
		default:
			return false
		}
	}

	SimpleService_EchoString_Helper.WrapResponse = func(success string, err error) (*SimpleService_EchoString_Result, error) {
		if err == nil {
			return &SimpleService_EchoString_Result{Success: &success}, nil
		}

		return nil, err
	}
	SimpleService_EchoString_Helper.UnwrapResponse = func(result *SimpleService_EchoString_Result) (success string, err error) {

		if result.Success != nil {
			success = *result.Success
			return
		}

		err = errors.New("expected a non-void result")
		return
	}

}

// SimpleService_EchoString_Result represents the result of a SimpleService.EchoString function call.
//
// The result of a EchoString execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type SimpleService_EchoString_Result struct {
	// Value returned by EchoString after a successful execution.
	Success *string `json:"success,omitempty"`
}

// ToWire translates a SimpleService_EchoString_Result struct into a Thrift-level intermediate
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
func (v *SimpleService_EchoString_Result) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Success != nil {
		w, err = wire.NewValueString(*(v.Success)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 0, Value: w}
		i++
	}

	if i != 1 {
		return wire.Value{}, fmt.Errorf("SimpleService_EchoString_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a SimpleService_EchoString_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a SimpleService_EchoString_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v SimpleService_EchoString_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *SimpleService_EchoString_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
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
		return fmt.Errorf("SimpleService_EchoString_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a SimpleService_EchoString_Result
// struct.
func (v *SimpleService_EchoString_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.Success != nil {
		fields[i] = fmt.Sprintf("Success: %v", *(v.Success))
		i++
	}

	return fmt.Sprintf("SimpleService_EchoString_Result{%v}", strings.Join(fields[:i], ", "))
}

func _String_EqualsPtr(lhs, rhs *string) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

// Equals returns true if all the fields of this SimpleService_EchoString_Result match the
// provided SimpleService_EchoString_Result.
//
// This function performs a deep comparison.
func (v *SimpleService_EchoString_Result) Equals(rhs *SimpleService_EchoString_Result) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !_String_EqualsPtr(v.Success, rhs.Success) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of SimpleService_EchoString_Result.
func (v *SimpleService_EchoString_Result) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.Success != nil {
		enc.AddString("success", *v.Success)
	}
	return err
}

// GetSuccess returns the value of Success if it is set or its
// zero value if it is unset.
func (v *SimpleService_EchoString_Result) GetSuccess() (o string) {
	if v != nil && v.Success != nil {
		return *v.Success
	}

	return
}

// IsSetSuccess returns true if Success is not nil.
func (v *SimpleService_EchoString_Result) IsSetSuccess() bool {
	return v != nil && v.Success != nil
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "EchoString" for this struct.
func (v *SimpleService_EchoString_Result) MethodName() string {
	return "EchoString"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *SimpleService_EchoString_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}