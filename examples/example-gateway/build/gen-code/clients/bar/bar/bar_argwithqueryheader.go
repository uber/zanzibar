// Code generated by thriftrw v1.9.0. DO NOT EDIT.
// @generated

package bar

import (
	"errors"
	"fmt"
	"go.uber.org/thriftrw/wire"
	"strings"
)

// Bar_ArgWithQueryHeader_Args represents the arguments for the Bar.argWithQueryHeader function.
//
// The arguments for argWithQueryHeader are sent and received over the wire as this struct.
type Bar_ArgWithQueryHeader_Args struct {
	UserUUID *string `json:"userUUID,omitempty"`
}

// ToWire translates a Bar_ArgWithQueryHeader_Args struct into a Thrift-level intermediate
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
func (v *Bar_ArgWithQueryHeader_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.UserUUID != nil {
		w, err = wire.NewValueString(*(v.UserUUID)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a Bar_ArgWithQueryHeader_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a Bar_ArgWithQueryHeader_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v Bar_ArgWithQueryHeader_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *Bar_ArgWithQueryHeader_Args) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.UserUUID = &x
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a Bar_ArgWithQueryHeader_Args
// struct.
func (v *Bar_ArgWithQueryHeader_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.UserUUID != nil {
		fields[i] = fmt.Sprintf("UserUUID: %v", *(v.UserUUID))
		i++
	}

	return fmt.Sprintf("Bar_ArgWithQueryHeader_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this Bar_ArgWithQueryHeader_Args match the
// provided Bar_ArgWithQueryHeader_Args.
//
// This function performs a deep comparison.
func (v *Bar_ArgWithQueryHeader_Args) Equals(rhs *Bar_ArgWithQueryHeader_Args) bool {
	if !_String_EqualsPtr(v.UserUUID, rhs.UserUUID) {
		return false
	}

	return true
}

// GetUserUUID returns the value of UserUUID if it is set or its
// zero value if it is unset.
func (v *Bar_ArgWithQueryHeader_Args) GetUserUUID() (o string) {
	if v.UserUUID != nil {
		return *v.UserUUID
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "argWithQueryHeader" for this struct.
func (v *Bar_ArgWithQueryHeader_Args) MethodName() string {
	return "argWithQueryHeader"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *Bar_ArgWithQueryHeader_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// Bar_ArgWithQueryHeader_Helper provides functions that aid in handling the
// parameters and return values of the Bar.argWithQueryHeader
// function.
var Bar_ArgWithQueryHeader_Helper = struct {
	// Args accepts the parameters of argWithQueryHeader in-order and returns
	// the arguments struct for the function.
	Args func(
		userUUID *string,
	) *Bar_ArgWithQueryHeader_Args

	// IsException returns true if the given error can be thrown
	// by argWithQueryHeader.
	//
	// An error can be thrown by argWithQueryHeader only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for argWithQueryHeader
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// argWithQueryHeader into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by argWithQueryHeader
	//
	//   value, err := argWithQueryHeader(args)
	//   result, err := Bar_ArgWithQueryHeader_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from argWithQueryHeader: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(*BarResponse, error) (*Bar_ArgWithQueryHeader_Result, error)

	// UnwrapResponse takes the result struct for argWithQueryHeader
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if argWithQueryHeader threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := Bar_ArgWithQueryHeader_Helper.UnwrapResponse(result)
	UnwrapResponse func(*Bar_ArgWithQueryHeader_Result) (*BarResponse, error)
}{}

func init() {
	Bar_ArgWithQueryHeader_Helper.Args = func(
		userUUID *string,
	) *Bar_ArgWithQueryHeader_Args {
		return &Bar_ArgWithQueryHeader_Args{
			UserUUID: userUUID,
		}
	}

	Bar_ArgWithQueryHeader_Helper.IsException = func(err error) bool {
		switch err.(type) {
		default:
			return false
		}
	}

	Bar_ArgWithQueryHeader_Helper.WrapResponse = func(success *BarResponse, err error) (*Bar_ArgWithQueryHeader_Result, error) {
		if err == nil {
			return &Bar_ArgWithQueryHeader_Result{Success: success}, nil
		}

		return nil, err
	}
	Bar_ArgWithQueryHeader_Helper.UnwrapResponse = func(result *Bar_ArgWithQueryHeader_Result) (success *BarResponse, err error) {

		if result.Success != nil {
			success = result.Success
			return
		}

		err = errors.New("expected a non-void result")
		return
	}

}

// Bar_ArgWithQueryHeader_Result represents the result of a Bar.argWithQueryHeader function call.
//
// The result of a argWithQueryHeader execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type Bar_ArgWithQueryHeader_Result struct {
	// Value returned by argWithQueryHeader after a successful execution.
	Success *BarResponse `json:"success,omitempty"`
}

// ToWire translates a Bar_ArgWithQueryHeader_Result struct into a Thrift-level intermediate
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
func (v *Bar_ArgWithQueryHeader_Result) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Success != nil {
		w, err = v.Success.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 0, Value: w}
		i++
	}

	if i != 1 {
		return wire.Value{}, fmt.Errorf("Bar_ArgWithQueryHeader_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a Bar_ArgWithQueryHeader_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a Bar_ArgWithQueryHeader_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v Bar_ArgWithQueryHeader_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *Bar_ArgWithQueryHeader_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TStruct {
				v.Success, err = _BarResponse_Read(field.Value)
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
		return fmt.Errorf("Bar_ArgWithQueryHeader_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a Bar_ArgWithQueryHeader_Result
// struct.
func (v *Bar_ArgWithQueryHeader_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.Success != nil {
		fields[i] = fmt.Sprintf("Success: %v", v.Success)
		i++
	}

	return fmt.Sprintf("Bar_ArgWithQueryHeader_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this Bar_ArgWithQueryHeader_Result match the
// provided Bar_ArgWithQueryHeader_Result.
//
// This function performs a deep comparison.
func (v *Bar_ArgWithQueryHeader_Result) Equals(rhs *Bar_ArgWithQueryHeader_Result) bool {
	if !((v.Success == nil && rhs.Success == nil) || (v.Success != nil && rhs.Success != nil && v.Success.Equals(rhs.Success))) {
		return false
	}

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "argWithQueryHeader" for this struct.
func (v *Bar_ArgWithQueryHeader_Result) MethodName() string {
	return "argWithQueryHeader"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *Bar_ArgWithQueryHeader_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
