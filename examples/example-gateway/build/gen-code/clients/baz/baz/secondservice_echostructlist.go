// Code generated by thriftrw v1.18.0. DO NOT EDIT.
// @generated

package baz

import (
	errors "errors"
	fmt "fmt"
	base "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/base"
	multierr "go.uber.org/multierr"
	wire "go.uber.org/thriftrw/wire"
	zapcore "go.uber.org/zap/zapcore"
	strings "strings"
)

// SecondService_EchoStructList_Args represents the arguments for the SecondService.echoStructList function.
//
// The arguments for echoStructList are sent and received over the wire as this struct.
type SecondService_EchoStructList_Args struct {
	Arg []*base.BazResponse `json:"arg,required"`
}

type _List_BazResponse_ValueList []*base.BazResponse

func (v _List_BazResponse_ValueList) ForEach(f func(wire.Value) error) error {
	for i, x := range v {
		if x == nil {
			return fmt.Errorf("invalid [%v]: value is nil", i)
		}
		w, err := x.ToWire()
		if err != nil {
			return err
		}
		err = f(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v _List_BazResponse_ValueList) Size() int {
	return len(v)
}

func (_List_BazResponse_ValueList) ValueType() wire.Type {
	return wire.TStruct
}

func (_List_BazResponse_ValueList) Close() {}

// ToWire translates a SecondService_EchoStructList_Args struct into a Thrift-level intermediate
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
func (v *SecondService_EchoStructList_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Arg == nil {
		return w, errors.New("field Arg of SecondService_EchoStructList_Args is required")
	}
	w, err = wire.NewValueList(_List_BazResponse_ValueList(v.Arg)), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _List_BazResponse_Read(l wire.ValueList) ([]*base.BazResponse, error) {
	if l.ValueType() != wire.TStruct {
		return nil, nil
	}

	o := make([]*base.BazResponse, 0, l.Size())
	err := l.ForEach(func(x wire.Value) error {
		i, err := _BazResponse_Read(x)
		if err != nil {
			return err
		}
		o = append(o, i)
		return nil
	})
	l.Close()
	return o, err
}

// FromWire deserializes a SecondService_EchoStructList_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a SecondService_EchoStructList_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v SecondService_EchoStructList_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *SecondService_EchoStructList_Args) FromWire(w wire.Value) error {
	var err error

	argIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TList {
				v.Arg, err = _List_BazResponse_Read(field.Value.GetList())
				if err != nil {
					return err
				}
				argIsSet = true
			}
		}
	}

	if !argIsSet {
		return errors.New("field Arg of SecondService_EchoStructList_Args is required")
	}

	return nil
}

// String returns a readable string representation of a SecondService_EchoStructList_Args
// struct.
func (v *SecondService_EchoStructList_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	fields[i] = fmt.Sprintf("Arg: %v", v.Arg)
	i++

	return fmt.Sprintf("SecondService_EchoStructList_Args{%v}", strings.Join(fields[:i], ", "))
}

func _List_BazResponse_Equals(lhs, rhs []*base.BazResponse) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i, lv := range lhs {
		rv := rhs[i]
		if !lv.Equals(rv) {
			return false
		}
	}

	return true
}

// Equals returns true if all the fields of this SecondService_EchoStructList_Args match the
// provided SecondService_EchoStructList_Args.
//
// This function performs a deep comparison.
func (v *SecondService_EchoStructList_Args) Equals(rhs *SecondService_EchoStructList_Args) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !_List_BazResponse_Equals(v.Arg, rhs.Arg) {
		return false
	}

	return true
}

type _List_BazResponse_Zapper []*base.BazResponse

// MarshalLogArray implements zapcore.ArrayMarshaler, enabling
// fast logging of _List_BazResponse_Zapper.
func (l _List_BazResponse_Zapper) MarshalLogArray(enc zapcore.ArrayEncoder) (err error) {
	for _, v := range l {
		err = multierr.Append(err, enc.AppendObject(v))
	}
	return err
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of SecondService_EchoStructList_Args.
func (v *SecondService_EchoStructList_Args) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	err = multierr.Append(err, enc.AddArray("arg", (_List_BazResponse_Zapper)(v.Arg)))
	return err
}

// GetArg returns the value of Arg if it is set or its
// zero value if it is unset.
func (v *SecondService_EchoStructList_Args) GetArg() (o []*base.BazResponse) {
	if v != nil {
		o = v.Arg
	}
	return
}

// IsSetArg returns true if Arg is not nil.
func (v *SecondService_EchoStructList_Args) IsSetArg() bool {
	return v != nil && v.Arg != nil
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "echoStructList" for this struct.
func (v *SecondService_EchoStructList_Args) MethodName() string {
	return "echoStructList"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *SecondService_EchoStructList_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// SecondService_EchoStructList_Helper provides functions that aid in handling the
// parameters and return values of the SecondService.echoStructList
// function.
var SecondService_EchoStructList_Helper = struct {
	// Args accepts the parameters of echoStructList in-order and returns
	// the arguments struct for the function.
	Args func(
		arg []*base.BazResponse,
	) *SecondService_EchoStructList_Args

	// IsException returns true if the given error can be thrown
	// by echoStructList.
	//
	// An error can be thrown by echoStructList only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for echoStructList
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// echoStructList into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by echoStructList
	//
	//   value, err := echoStructList(args)
	//   result, err := SecondService_EchoStructList_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from echoStructList: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func([]*base.BazResponse, error) (*SecondService_EchoStructList_Result, error)

	// UnwrapResponse takes the result struct for echoStructList
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if echoStructList threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := SecondService_EchoStructList_Helper.UnwrapResponse(result)
	UnwrapResponse func(*SecondService_EchoStructList_Result) ([]*base.BazResponse, error)
}{}

func init() {
	SecondService_EchoStructList_Helper.Args = func(
		arg []*base.BazResponse,
	) *SecondService_EchoStructList_Args {
		return &SecondService_EchoStructList_Args{
			Arg: arg,
		}
	}

	SecondService_EchoStructList_Helper.IsException = func(err error) bool {
		switch err.(type) {
		default:
			return false
		}
	}

	SecondService_EchoStructList_Helper.WrapResponse = func(success []*base.BazResponse, err error) (*SecondService_EchoStructList_Result, error) {
		if err == nil {
			return &SecondService_EchoStructList_Result{Success: success}, nil
		}

		return nil, err
	}
	SecondService_EchoStructList_Helper.UnwrapResponse = func(result *SecondService_EchoStructList_Result) (success []*base.BazResponse, err error) {

		if result.Success != nil {
			success = result.Success
			return
		}

		err = errors.New("expected a non-void result")
		return
	}

}

// SecondService_EchoStructList_Result represents the result of a SecondService.echoStructList function call.
//
// The result of a echoStructList execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type SecondService_EchoStructList_Result struct {
	// Value returned by echoStructList after a successful execution.
	Success []*base.BazResponse `json:"success,omitempty"`
}

// ToWire translates a SecondService_EchoStructList_Result struct into a Thrift-level intermediate
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
func (v *SecondService_EchoStructList_Result) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Success != nil {
		w, err = wire.NewValueList(_List_BazResponse_ValueList(v.Success)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 0, Value: w}
		i++
	}

	if i != 1 {
		return wire.Value{}, fmt.Errorf("SecondService_EchoStructList_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a SecondService_EchoStructList_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a SecondService_EchoStructList_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v SecondService_EchoStructList_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *SecondService_EchoStructList_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TList {
				v.Success, err = _List_BazResponse_Read(field.Value.GetList())
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
		return fmt.Errorf("SecondService_EchoStructList_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a SecondService_EchoStructList_Result
// struct.
func (v *SecondService_EchoStructList_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.Success != nil {
		fields[i] = fmt.Sprintf("Success: %v", v.Success)
		i++
	}

	return fmt.Sprintf("SecondService_EchoStructList_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this SecondService_EchoStructList_Result match the
// provided SecondService_EchoStructList_Result.
//
// This function performs a deep comparison.
func (v *SecondService_EchoStructList_Result) Equals(rhs *SecondService_EchoStructList_Result) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !((v.Success == nil && rhs.Success == nil) || (v.Success != nil && rhs.Success != nil && _List_BazResponse_Equals(v.Success, rhs.Success))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of SecondService_EchoStructList_Result.
func (v *SecondService_EchoStructList_Result) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.Success != nil {
		err = multierr.Append(err, enc.AddArray("success", (_List_BazResponse_Zapper)(v.Success)))
	}
	return err
}

// GetSuccess returns the value of Success if it is set or its
// zero value if it is unset.
func (v *SecondService_EchoStructList_Result) GetSuccess() (o []*base.BazResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}

	return
}

// IsSetSuccess returns true if Success is not nil.
func (v *SecondService_EchoStructList_Result) IsSetSuccess() bool {
	return v != nil && v.Success != nil
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "echoStructList" for this struct.
func (v *SecondService_EchoStructList_Result) MethodName() string {
	return "echoStructList"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *SecondService_EchoStructList_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
