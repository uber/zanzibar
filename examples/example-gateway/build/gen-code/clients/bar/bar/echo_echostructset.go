// Code generated by thriftrw v1.8.0. DO NOT EDIT.
// @generated

package bar

import (
	"errors"
	"fmt"
	"go.uber.org/thriftrw/wire"
	"strings"
)

// Echo_EchoStructSet_Args represents the arguments for the Echo.echoStructSet function.
//
// The arguments for echoStructSet are sent and received over the wire as this struct.
type Echo_EchoStructSet_Args struct {
	Arg []*BarResponse `json:"arg,required"`
}

type _Set_BarResponse_ValueList []*BarResponse

func (v _Set_BarResponse_ValueList) ForEach(f func(wire.Value) error) error {
	for _, x := range v {
		if x == nil {
			return fmt.Errorf("invalid set item: value is nil")
		}
		w, err := x.ToWire()
		if err != nil {
			return err
		}

		if err := f(w); err != nil {
			return err
		}
	}
	return nil
}

func (v _Set_BarResponse_ValueList) Size() int {
	return len(v)
}

func (_Set_BarResponse_ValueList) ValueType() wire.Type {
	return wire.TStruct
}

func (_Set_BarResponse_ValueList) Close() {}

// ToWire translates a Echo_EchoStructSet_Args struct into a Thrift-level intermediate
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
func (v *Echo_EchoStructSet_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Arg == nil {
		return w, errors.New("field Arg of Echo_EchoStructSet_Args is required")
	}
	w, err = wire.NewValueSet(_Set_BarResponse_ValueList(v.Arg)), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _Set_BarResponse_Read(s wire.ValueList) ([]*BarResponse, error) {
	if s.ValueType() != wire.TStruct {
		return nil, nil
	}

	o := make([]*BarResponse, 0, s.Size())
	err := s.ForEach(func(x wire.Value) error {
		i, err := _BarResponse_Read(x)
		if err != nil {
			return err
		}

		o = append(o, i)
		return nil
	})
	s.Close()
	return o, err
}

// FromWire deserializes a Echo_EchoStructSet_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a Echo_EchoStructSet_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v Echo_EchoStructSet_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *Echo_EchoStructSet_Args) FromWire(w wire.Value) error {
	var err error

	argIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TSet {
				v.Arg, err = _Set_BarResponse_Read(field.Value.GetSet())
				if err != nil {
					return err
				}
				argIsSet = true
			}
		}
	}

	if !argIsSet {
		return errors.New("field Arg of Echo_EchoStructSet_Args is required")
	}

	return nil
}

// String returns a readable string representation of a Echo_EchoStructSet_Args
// struct.
func (v *Echo_EchoStructSet_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	fields[i] = fmt.Sprintf("Arg: %v", v.Arg)
	i++

	return fmt.Sprintf("Echo_EchoStructSet_Args{%v}", strings.Join(fields[:i], ", "))
}

func _Set_BarResponse_Equals(lhs, rhs []*BarResponse) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for _, x := range lhs {
		ok := false
		for _, y := range rhs {
			if x.Equals(y) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}

	return true
}

// Equals returns true if all the fields of this Echo_EchoStructSet_Args match the
// provided Echo_EchoStructSet_Args.
//
// This function performs a deep comparison.
func (v *Echo_EchoStructSet_Args) Equals(rhs *Echo_EchoStructSet_Args) bool {
	if !_Set_BarResponse_Equals(v.Arg, rhs.Arg) {
		return false
	}

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "echoStructSet" for this struct.
func (v *Echo_EchoStructSet_Args) MethodName() string {
	return "echoStructSet"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *Echo_EchoStructSet_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// Echo_EchoStructSet_Helper provides functions that aid in handling the
// parameters and return values of the Echo.echoStructSet
// function.
var Echo_EchoStructSet_Helper = struct {
	// Args accepts the parameters of echoStructSet in-order and returns
	// the arguments struct for the function.
	Args func(
		arg []*BarResponse,
	) *Echo_EchoStructSet_Args

	// IsException returns true if the given error can be thrown
	// by echoStructSet.
	//
	// An error can be thrown by echoStructSet only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for echoStructSet
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// echoStructSet into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by echoStructSet
	//
	//   value, err := echoStructSet(args)
	//   result, err := Echo_EchoStructSet_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from echoStructSet: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func([]*BarResponse, error) (*Echo_EchoStructSet_Result, error)

	// UnwrapResponse takes the result struct for echoStructSet
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if echoStructSet threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := Echo_EchoStructSet_Helper.UnwrapResponse(result)
	UnwrapResponse func(*Echo_EchoStructSet_Result) ([]*BarResponse, error)
}{}

func init() {
	Echo_EchoStructSet_Helper.Args = func(
		arg []*BarResponse,
	) *Echo_EchoStructSet_Args {
		return &Echo_EchoStructSet_Args{
			Arg: arg,
		}
	}

	Echo_EchoStructSet_Helper.IsException = func(err error) bool {
		switch err.(type) {
		default:
			return false
		}
	}

	Echo_EchoStructSet_Helper.WrapResponse = func(success []*BarResponse, err error) (*Echo_EchoStructSet_Result, error) {
		if err == nil {
			return &Echo_EchoStructSet_Result{Success: success}, nil
		}

		return nil, err
	}
	Echo_EchoStructSet_Helper.UnwrapResponse = func(result *Echo_EchoStructSet_Result) (success []*BarResponse, err error) {

		if result.Success != nil {
			success = result.Success
			return
		}

		err = errors.New("expected a non-void result")
		return
	}

}

// Echo_EchoStructSet_Result represents the result of a Echo.echoStructSet function call.
//
// The result of a echoStructSet execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type Echo_EchoStructSet_Result struct {
	// Value returned by echoStructSet after a successful execution.
	Success []*BarResponse `json:"success,omitempty"`
}

// ToWire translates a Echo_EchoStructSet_Result struct into a Thrift-level intermediate
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
func (v *Echo_EchoStructSet_Result) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Success != nil {
		w, err = wire.NewValueSet(_Set_BarResponse_ValueList(v.Success)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 0, Value: w}
		i++
	}

	if i != 1 {
		return wire.Value{}, fmt.Errorf("Echo_EchoStructSet_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a Echo_EchoStructSet_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a Echo_EchoStructSet_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v Echo_EchoStructSet_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *Echo_EchoStructSet_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TSet {
				v.Success, err = _Set_BarResponse_Read(field.Value.GetSet())
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
		return fmt.Errorf("Echo_EchoStructSet_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a Echo_EchoStructSet_Result
// struct.
func (v *Echo_EchoStructSet_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.Success != nil {
		fields[i] = fmt.Sprintf("Success: %v", v.Success)
		i++
	}

	return fmt.Sprintf("Echo_EchoStructSet_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this Echo_EchoStructSet_Result match the
// provided Echo_EchoStructSet_Result.
//
// This function performs a deep comparison.
func (v *Echo_EchoStructSet_Result) Equals(rhs *Echo_EchoStructSet_Result) bool {
	if !((v.Success == nil && rhs.Success == nil) || (v.Success != nil && rhs.Success != nil && _Set_BarResponse_Equals(v.Success, rhs.Success))) {
		return false
	}

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "echoStructSet" for this struct.
func (v *Echo_EchoStructSet_Result) MethodName() string {
	return "echoStructSet"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *Echo_EchoStructSet_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
