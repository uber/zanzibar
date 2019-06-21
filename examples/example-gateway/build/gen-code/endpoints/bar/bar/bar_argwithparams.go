// Code generated by thriftrw v1.18.0. DO NOT EDIT.
// @generated

package bar

import (
	errors "errors"
	fmt "fmt"
	multierr "go.uber.org/multierr"
	wire "go.uber.org/thriftrw/wire"
	zapcore "go.uber.org/zap/zapcore"
	strings "strings"
)

// Bar_ArgWithParams_Args represents the arguments for the Bar.argWithParams function.
//
// The arguments for argWithParams are sent and received over the wire as this struct.
type Bar_ArgWithParams_Args struct {
	UUID   string        `json:"-"`
	Params *ParamsStruct `json:"params,omitempty"`
}

// ToWire translates a Bar_ArgWithParams_Args struct into a Thrift-level intermediate
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
func (v *Bar_ArgWithParams_Args) ToWire() (wire.Value, error) {
	var (
		fields [2]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueString(v.UUID), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++
	if v.Params != nil {
		w, err = v.Params.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 2, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _ParamsStruct_Read(w wire.Value) (*ParamsStruct, error) {
	var v ParamsStruct
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a Bar_ArgWithParams_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a Bar_ArgWithParams_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v Bar_ArgWithParams_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *Bar_ArgWithParams_Args) FromWire(w wire.Value) error {
	var err error

	uuidIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.UUID, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				uuidIsSet = true
			}
		case 2:
			if field.Value.Type() == wire.TStruct {
				v.Params, err = _ParamsStruct_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	if !uuidIsSet {
		return errors.New("field UUID of Bar_ArgWithParams_Args is required")
	}

	return nil
}

// String returns a readable string representation of a Bar_ArgWithParams_Args
// struct.
func (v *Bar_ArgWithParams_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [2]string
	i := 0
	fields[i] = fmt.Sprintf("UUID: %v", v.UUID)
	i++
	if v.Params != nil {
		fields[i] = fmt.Sprintf("Params: %v", v.Params)
		i++
	}

	return fmt.Sprintf("Bar_ArgWithParams_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this Bar_ArgWithParams_Args match the
// provided Bar_ArgWithParams_Args.
//
// This function performs a deep comparison.
func (v *Bar_ArgWithParams_Args) Equals(rhs *Bar_ArgWithParams_Args) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !(v.UUID == rhs.UUID) {
		return false
	}
	if !((v.Params == nil && rhs.Params == nil) || (v.Params != nil && rhs.Params != nil && v.Params.Equals(rhs.Params))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of Bar_ArgWithParams_Args.
func (v *Bar_ArgWithParams_Args) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	enc.AddString("uuid", v.UUID)
	if v.Params != nil {
		err = multierr.Append(err, enc.AddObject("params", v.Params))
	}
	return err
}

// GetUUID returns the value of UUID if it is set or its
// zero value if it is unset.
func (v *Bar_ArgWithParams_Args) GetUUID() (o string) {
	if v != nil {
		o = v.UUID
	}
	return
}

// GetParams returns the value of Params if it is set or its
// zero value if it is unset.
func (v *Bar_ArgWithParams_Args) GetParams() (o *ParamsStruct) {
	if v != nil && v.Params != nil {
		return v.Params
	}

	return
}

// IsSetParams returns true if Params is not nil.
func (v *Bar_ArgWithParams_Args) IsSetParams() bool {
	return v != nil && v.Params != nil
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "argWithParams" for this struct.
func (v *Bar_ArgWithParams_Args) MethodName() string {
	return "argWithParams"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *Bar_ArgWithParams_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// Bar_ArgWithParams_Helper provides functions that aid in handling the
// parameters and return values of the Bar.argWithParams
// function.
var Bar_ArgWithParams_Helper = struct {
	// Args accepts the parameters of argWithParams in-order and returns
	// the arguments struct for the function.
	Args func(
		uuid string,
		params *ParamsStruct,
	) *Bar_ArgWithParams_Args

	// IsException returns true if the given error can be thrown
	// by argWithParams.
	//
	// An error can be thrown by argWithParams only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for argWithParams
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// argWithParams into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by argWithParams
	//
	//   value, err := argWithParams(args)
	//   result, err := Bar_ArgWithParams_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from argWithParams: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(*BarResponse, error) (*Bar_ArgWithParams_Result, error)

	// UnwrapResponse takes the result struct for argWithParams
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if argWithParams threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := Bar_ArgWithParams_Helper.UnwrapResponse(result)
	UnwrapResponse func(*Bar_ArgWithParams_Result) (*BarResponse, error)
}{}

func init() {
	Bar_ArgWithParams_Helper.Args = func(
		uuid string,
		params *ParamsStruct,
	) *Bar_ArgWithParams_Args {
		return &Bar_ArgWithParams_Args{
			UUID:   uuid,
			Params: params,
		}
	}

	Bar_ArgWithParams_Helper.IsException = func(err error) bool {
		switch err.(type) {
		default:
			return false
		}
	}

	Bar_ArgWithParams_Helper.WrapResponse = func(success *BarResponse, err error) (*Bar_ArgWithParams_Result, error) {
		if err == nil {
			return &Bar_ArgWithParams_Result{Success: success}, nil
		}

		return nil, err
	}
	Bar_ArgWithParams_Helper.UnwrapResponse = func(result *Bar_ArgWithParams_Result) (success *BarResponse, err error) {

		if result.Success != nil {
			success = result.Success
			return
		}

		err = errors.New("expected a non-void result")
		return
	}

}

// Bar_ArgWithParams_Result represents the result of a Bar.argWithParams function call.
//
// The result of a argWithParams execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type Bar_ArgWithParams_Result struct {
	// Value returned by argWithParams after a successful execution.
	Success *BarResponse `json:"success,omitempty"`
}

// ToWire translates a Bar_ArgWithParams_Result struct into a Thrift-level intermediate
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
func (v *Bar_ArgWithParams_Result) ToWire() (wire.Value, error) {
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
		return wire.Value{}, fmt.Errorf("Bar_ArgWithParams_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a Bar_ArgWithParams_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a Bar_ArgWithParams_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v Bar_ArgWithParams_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *Bar_ArgWithParams_Result) FromWire(w wire.Value) error {
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
		return fmt.Errorf("Bar_ArgWithParams_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a Bar_ArgWithParams_Result
// struct.
func (v *Bar_ArgWithParams_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.Success != nil {
		fields[i] = fmt.Sprintf("Success: %v", v.Success)
		i++
	}

	return fmt.Sprintf("Bar_ArgWithParams_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this Bar_ArgWithParams_Result match the
// provided Bar_ArgWithParams_Result.
//
// This function performs a deep comparison.
func (v *Bar_ArgWithParams_Result) Equals(rhs *Bar_ArgWithParams_Result) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !((v.Success == nil && rhs.Success == nil) || (v.Success != nil && rhs.Success != nil && v.Success.Equals(rhs.Success))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of Bar_ArgWithParams_Result.
func (v *Bar_ArgWithParams_Result) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.Success != nil {
		err = multierr.Append(err, enc.AddObject("success", v.Success))
	}
	return err
}

// GetSuccess returns the value of Success if it is set or its
// zero value if it is unset.
func (v *Bar_ArgWithParams_Result) GetSuccess() (o *BarResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}

	return
}

// IsSetSuccess returns true if Success is not nil.
func (v *Bar_ArgWithParams_Result) IsSetSuccess() bool {
	return v != nil && v.Success != nil
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "argWithParams" for this struct.
func (v *Bar_ArgWithParams_Result) MethodName() string {
	return "argWithParams"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *Bar_ArgWithParams_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
