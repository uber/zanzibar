// Code generated by thriftrw v1.8.0. DO NOT EDIT.
// @generated

package bar

import (
	"errors"
	"fmt"
	"go.uber.org/thriftrw/wire"
	"strings"
)

// Bar_ArgWithQueryParams_Args represents the arguments for the Bar.argWithQueryParams function.
//
// The arguments for argWithQueryParams are sent and received over the wire as this struct.
type Bar_ArgWithQueryParams_Args struct {
	Name     string   `json:"name,required"`
	UserUUID *string  `json:"userUUID,omitempty"`
	Foo      []string `json:"foo,omitempty"`
	Bar      []int8   `json:"bar,required"`
}

type _List_Byte_ValueList []int8

func (v _List_Byte_ValueList) ForEach(f func(wire.Value) error) error {
	for _, x := range v {
		w, err := wire.NewValueI8(x), error(nil)
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

func (v _List_Byte_ValueList) Size() int {
	return len(v)
}

func (_List_Byte_ValueList) ValueType() wire.Type {
	return wire.TI8
}

func (_List_Byte_ValueList) Close() {}

// ToWire translates a Bar_ArgWithQueryParams_Args struct into a Thrift-level intermediate
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
func (v *Bar_ArgWithQueryParams_Args) ToWire() (wire.Value, error) {
	var (
		fields [4]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueString(v.Name), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++
	if v.UserUUID != nil {
		w, err = wire.NewValueString(*(v.UserUUID)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 2, Value: w}
		i++
	}
	if v.Foo != nil {
		w, err = wire.NewValueList(_List_String_ValueList(v.Foo)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 3, Value: w}
		i++
	}
	if v.Bar == nil {
		return w, errors.New("field Bar of Bar_ArgWithQueryParams_Args is required")
	}
	w, err = wire.NewValueList(_List_Byte_ValueList(v.Bar)), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 4, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _List_Byte_Read(l wire.ValueList) ([]int8, error) {
	if l.ValueType() != wire.TI8 {
		return nil, nil
	}

	o := make([]int8, 0, l.Size())
	err := l.ForEach(func(x wire.Value) error {
		i, err := x.GetI8(), error(nil)
		if err != nil {
			return err
		}
		o = append(o, i)
		return nil
	})
	l.Close()
	return o, err
}

// FromWire deserializes a Bar_ArgWithQueryParams_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a Bar_ArgWithQueryParams_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v Bar_ArgWithQueryParams_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *Bar_ArgWithQueryParams_Args) FromWire(w wire.Value) error {
	var err error

	nameIsSet := false

	barIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.Name, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				nameIsSet = true
			}
		case 2:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.UserUUID = &x
				if err != nil {
					return err
				}

			}
		case 3:
			if field.Value.Type() == wire.TList {
				v.Foo, err = _List_String_Read(field.Value.GetList())
				if err != nil {
					return err
				}

			}
		case 4:
			if field.Value.Type() == wire.TList {
				v.Bar, err = _List_Byte_Read(field.Value.GetList())
				if err != nil {
					return err
				}
				barIsSet = true
			}
		}
	}

	if !nameIsSet {
		return errors.New("field Name of Bar_ArgWithQueryParams_Args is required")
	}

	if !barIsSet {
		return errors.New("field Bar of Bar_ArgWithQueryParams_Args is required")
	}

	return nil
}

// String returns a readable string representation of a Bar_ArgWithQueryParams_Args
// struct.
func (v *Bar_ArgWithQueryParams_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [4]string
	i := 0
	fields[i] = fmt.Sprintf("Name: %v", v.Name)
	i++
	if v.UserUUID != nil {
		fields[i] = fmt.Sprintf("UserUUID: %v", *(v.UserUUID))
		i++
	}
	if v.Foo != nil {
		fields[i] = fmt.Sprintf("Foo: %v", v.Foo)
		i++
	}
	fields[i] = fmt.Sprintf("Bar: %v", v.Bar)
	i++

	return fmt.Sprintf("Bar_ArgWithQueryParams_Args{%v}", strings.Join(fields[:i], ", "))
}

func _List_Byte_Equals(lhs, rhs []int8) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i, lv := range lhs {
		rv := rhs[i]
		if !(lv == rv) {
			return false
		}
	}

	return true
}

// Equals returns true if all the fields of this Bar_ArgWithQueryParams_Args match the
// provided Bar_ArgWithQueryParams_Args.
//
// This function performs a deep comparison.
func (v *Bar_ArgWithQueryParams_Args) Equals(rhs *Bar_ArgWithQueryParams_Args) bool {
	if !(v.Name == rhs.Name) {
		return false
	}
	if !_String_EqualsPtr(v.UserUUID, rhs.UserUUID) {
		return false
	}
	if !((v.Foo == nil && rhs.Foo == nil) || (v.Foo != nil && rhs.Foo != nil && _List_String_Equals(v.Foo, rhs.Foo))) {
		return false
	}
	if !_List_Byte_Equals(v.Bar, rhs.Bar) {
		return false
	}

	return true
}

// GetUserUUID returns the value of UserUUID if it is set or its
// zero value if it is unset.
func (v *Bar_ArgWithQueryParams_Args) GetUserUUID() (o string) {
	if v.UserUUID != nil {
		return *v.UserUUID
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "argWithQueryParams" for this struct.
func (v *Bar_ArgWithQueryParams_Args) MethodName() string {
	return "argWithQueryParams"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *Bar_ArgWithQueryParams_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// Bar_ArgWithQueryParams_Helper provides functions that aid in handling the
// parameters and return values of the Bar.argWithQueryParams
// function.
var Bar_ArgWithQueryParams_Helper = struct {
	// Args accepts the parameters of argWithQueryParams in-order and returns
	// the arguments struct for the function.
	Args func(
		name string,
		userUUID *string,
		foo []string,
		bar []int8,
	) *Bar_ArgWithQueryParams_Args

	// IsException returns true if the given error can be thrown
	// by argWithQueryParams.
	//
	// An error can be thrown by argWithQueryParams only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for argWithQueryParams
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// argWithQueryParams into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by argWithQueryParams
	//
	//   value, err := argWithQueryParams(args)
	//   result, err := Bar_ArgWithQueryParams_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from argWithQueryParams: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(*BarResponse, error) (*Bar_ArgWithQueryParams_Result, error)

	// UnwrapResponse takes the result struct for argWithQueryParams
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if argWithQueryParams threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := Bar_ArgWithQueryParams_Helper.UnwrapResponse(result)
	UnwrapResponse func(*Bar_ArgWithQueryParams_Result) (*BarResponse, error)
}{}

func init() {
	Bar_ArgWithQueryParams_Helper.Args = func(
		name string,
		userUUID *string,
		foo []string,
		bar []int8,
	) *Bar_ArgWithQueryParams_Args {
		return &Bar_ArgWithQueryParams_Args{
			Name:     name,
			UserUUID: userUUID,
			Foo:      foo,
			Bar:      bar,
		}
	}

	Bar_ArgWithQueryParams_Helper.IsException = func(err error) bool {
		switch err.(type) {
		default:
			return false
		}
	}

	Bar_ArgWithQueryParams_Helper.WrapResponse = func(success *BarResponse, err error) (*Bar_ArgWithQueryParams_Result, error) {
		if err == nil {
			return &Bar_ArgWithQueryParams_Result{Success: success}, nil
		}

		return nil, err
	}
	Bar_ArgWithQueryParams_Helper.UnwrapResponse = func(result *Bar_ArgWithQueryParams_Result) (success *BarResponse, err error) {

		if result.Success != nil {
			success = result.Success
			return
		}

		err = errors.New("expected a non-void result")
		return
	}

}

// Bar_ArgWithQueryParams_Result represents the result of a Bar.argWithQueryParams function call.
//
// The result of a argWithQueryParams execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type Bar_ArgWithQueryParams_Result struct {
	// Value returned by argWithQueryParams after a successful execution.
	Success *BarResponse `json:"success,omitempty"`
}

// ToWire translates a Bar_ArgWithQueryParams_Result struct into a Thrift-level intermediate
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
func (v *Bar_ArgWithQueryParams_Result) ToWire() (wire.Value, error) {
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
		return wire.Value{}, fmt.Errorf("Bar_ArgWithQueryParams_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a Bar_ArgWithQueryParams_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a Bar_ArgWithQueryParams_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v Bar_ArgWithQueryParams_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *Bar_ArgWithQueryParams_Result) FromWire(w wire.Value) error {
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
		return fmt.Errorf("Bar_ArgWithQueryParams_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a Bar_ArgWithQueryParams_Result
// struct.
func (v *Bar_ArgWithQueryParams_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.Success != nil {
		fields[i] = fmt.Sprintf("Success: %v", v.Success)
		i++
	}

	return fmt.Sprintf("Bar_ArgWithQueryParams_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this Bar_ArgWithQueryParams_Result match the
// provided Bar_ArgWithQueryParams_Result.
//
// This function performs a deep comparison.
func (v *Bar_ArgWithQueryParams_Result) Equals(rhs *Bar_ArgWithQueryParams_Result) bool {
	if !((v.Success == nil && rhs.Success == nil) || (v.Success != nil && rhs.Success != nil && v.Success.Equals(rhs.Success))) {
		return false
	}

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "argWithQueryParams" for this struct.
func (v *Bar_ArgWithQueryParams_Result) MethodName() string {
	return "argWithQueryParams"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *Bar_ArgWithQueryParams_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
