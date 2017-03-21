// Code generated by thriftrw v1.0.0
// @generated

package baz

import (
	"errors"
	"fmt"
	"go.uber.org/thriftrw/wire"
	"strings"
)

type SimpleService_Call_Args struct {
	Arg *Data `json:"arg,omitempty"`
}

func (v *SimpleService_Call_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)
	if v.Arg != nil {
		w, err = v.Arg.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}
	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _Data_Read(w wire.Value) (*Data, error) {
	var v Data
	err := v.FromWire(w)
	return &v, err
}

func (v *SimpleService_Call_Args) FromWire(w wire.Value) error {
	var err error
	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.Arg, err = _Data_Read(field.Value)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (v *SimpleService_Call_Args) String() string {
	var fields [1]string
	i := 0
	if v.Arg != nil {
		fields[i] = fmt.Sprintf("Arg: %v", v.Arg)
		i++
	}
	return fmt.Sprintf("SimpleService_Call_Args{%v}", strings.Join(fields[:i], ", "))
}

func (v *SimpleService_Call_Args) MethodName() string {
	return "Call"
}

func (v *SimpleService_Call_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

var SimpleService_Call_Helper = struct {
	Args           func(arg *Data) *SimpleService_Call_Args
	IsException    func(error) bool
	WrapResponse   func(*Data, error) (*SimpleService_Call_Result, error)
	UnwrapResponse func(*SimpleService_Call_Result) (*Data, error)
}{}

func init() {
	SimpleService_Call_Helper.Args = func(arg *Data) *SimpleService_Call_Args {
		return &SimpleService_Call_Args{Arg: arg}
	}
	SimpleService_Call_Helper.IsException = func(err error) bool {
		switch err.(type) {
		default:
			return false
		}
	}
	SimpleService_Call_Helper.WrapResponse = func(success *Data, err error) (*SimpleService_Call_Result, error) {
		if err == nil {
			return &SimpleService_Call_Result{Success: success}, nil
		}
		return nil, err
	}
	SimpleService_Call_Helper.UnwrapResponse = func(result *SimpleService_Call_Result) (success *Data, err error) {
		if result.Success != nil {
			success = result.Success
			return
		}
		err = errors.New("expected a non-void result")
		return
	}
}

type SimpleService_Call_Result struct {
	Success *Data `json:"success,omitempty"`
}

func (v *SimpleService_Call_Result) ToWire() (wire.Value, error) {
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
		return wire.Value{}, fmt.Errorf("SimpleService_Call_Result should have exactly one field: got %v fields", i)
	}
	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func (v *SimpleService_Call_Result) FromWire(w wire.Value) error {
	var err error
	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TStruct {
				v.Success, err = _Data_Read(field.Value)
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
		return fmt.Errorf("SimpleService_Call_Result should have exactly one field: got %v fields", count)
	}
	return nil
}

func (v *SimpleService_Call_Result) String() string {
	var fields [1]string
	i := 0
	if v.Success != nil {
		fields[i] = fmt.Sprintf("Success: %v", v.Success)
		i++
	}
	return fmt.Sprintf("SimpleService_Call_Result{%v}", strings.Join(fields[:i], ", "))
}

func (v *SimpleService_Call_Result) MethodName() string {
	return "Call"
}

func (v *SimpleService_Call_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
