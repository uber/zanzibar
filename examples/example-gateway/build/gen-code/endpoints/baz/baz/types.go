// Code generated by thriftrw v1.8.0. DO NOT EDIT.
// @generated

package baz

import (
	"errors"
	"fmt"
	"go.uber.org/thriftrw/wire"
	"strings"
)

type AuthErr struct {
	Message string `json:"message,required"`
}

// ToWire translates a AuthErr struct into a Thrift-level intermediate
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
func (v *AuthErr) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueString(v.Message), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a AuthErr struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a AuthErr struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v AuthErr
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *AuthErr) FromWire(w wire.Value) error {
	var err error

	messageIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.Message, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				messageIsSet = true
			}
		}
	}

	if !messageIsSet {
		return errors.New("field Message of AuthErr is required")
	}

	return nil
}

// String returns a readable string representation of a AuthErr
// struct.
func (v *AuthErr) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	fields[i] = fmt.Sprintf("Message: %v", v.Message)
	i++

	return fmt.Sprintf("AuthErr{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this AuthErr match the
// provided AuthErr.
//
// This function performs a deep comparison.
func (v *AuthErr) Equals(rhs *AuthErr) bool {
	if !(v.Message == rhs.Message) {
		return false
	}

	return true
}

func (v *AuthErr) Error() string {
	return v.String()
}

type BazRequest struct {
	B1 bool   `json:"b1,required"`
	S2 string `json:"s2,required"`
	I3 int32  `json:"i3,required"`
}

// ToWire translates a BazRequest struct into a Thrift-level intermediate
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
func (v *BazRequest) ToWire() (wire.Value, error) {
	var (
		fields [3]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueBool(v.B1), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++

	w, err = wire.NewValueString(v.S2), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 2, Value: w}
	i++

	w, err = wire.NewValueI32(v.I3), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 3, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a BazRequest struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a BazRequest struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v BazRequest
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *BazRequest) FromWire(w wire.Value) error {
	var err error

	b1IsSet := false
	s2IsSet := false
	i3IsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBool {
				v.B1, err = field.Value.GetBool(), error(nil)
				if err != nil {
					return err
				}
				b1IsSet = true
			}
		case 2:
			if field.Value.Type() == wire.TBinary {
				v.S2, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				s2IsSet = true
			}
		case 3:
			if field.Value.Type() == wire.TI32 {
				v.I3, err = field.Value.GetI32(), error(nil)
				if err != nil {
					return err
				}
				i3IsSet = true
			}
		}
	}

	if !b1IsSet {
		return errors.New("field B1 of BazRequest is required")
	}

	if !s2IsSet {
		return errors.New("field S2 of BazRequest is required")
	}

	if !i3IsSet {
		return errors.New("field I3 of BazRequest is required")
	}

	return nil
}

// String returns a readable string representation of a BazRequest
// struct.
func (v *BazRequest) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [3]string
	i := 0
	fields[i] = fmt.Sprintf("B1: %v", v.B1)
	i++
	fields[i] = fmt.Sprintf("S2: %v", v.S2)
	i++
	fields[i] = fmt.Sprintf("I3: %v", v.I3)
	i++

	return fmt.Sprintf("BazRequest{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this BazRequest match the
// provided BazRequest.
//
// This function performs a deep comparison.
func (v *BazRequest) Equals(rhs *BazRequest) bool {
	if !(v.B1 == rhs.B1) {
		return false
	}
	if !(v.S2 == rhs.S2) {
		return false
	}
	if !(v.I3 == rhs.I3) {
		return false
	}

	return true
}

type BazResponse struct {
	Message string `json:"message,required"`
}

// ToWire translates a BazResponse struct into a Thrift-level intermediate
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
func (v *BazResponse) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueString(v.Message), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a BazResponse struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a BazResponse struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v BazResponse
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *BazResponse) FromWire(w wire.Value) error {
	var err error

	messageIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.Message, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				messageIsSet = true
			}
		}
	}

	if !messageIsSet {
		return errors.New("field Message of BazResponse is required")
	}

	return nil
}

// String returns a readable string representation of a BazResponse
// struct.
func (v *BazResponse) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	fields[i] = fmt.Sprintf("Message: %v", v.Message)
	i++

	return fmt.Sprintf("BazResponse{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this BazResponse match the
// provided BazResponse.
//
// This function performs a deep comparison.
func (v *BazResponse) Equals(rhs *BazResponse) bool {
	if !(v.Message == rhs.Message) {
		return false
	}

	return true
}

type HeaderSchema struct {
}

// ToWire translates a HeaderSchema struct into a Thrift-level intermediate
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
func (v *HeaderSchema) ToWire() (wire.Value, error) {
	var (
		fields [0]wire.Field
		i      int = 0
	)

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a HeaderSchema struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a HeaderSchema struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v HeaderSchema
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *HeaderSchema) FromWire(w wire.Value) error {

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		}
	}

	return nil
}

// String returns a readable string representation of a HeaderSchema
// struct.
func (v *HeaderSchema) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [0]string
	i := 0

	return fmt.Sprintf("HeaderSchema{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this HeaderSchema match the
// provided HeaderSchema.
//
// This function performs a deep comparison.
func (v *HeaderSchema) Equals(rhs *HeaderSchema) bool {

	return true
}

type NestedStruct struct {
	Msg   string `json:"msg,required"`
	Check *int32 `json:"check,omitempty"`
}

// ToWire translates a NestedStruct struct into a Thrift-level intermediate
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
func (v *NestedStruct) ToWire() (wire.Value, error) {
	var (
		fields [2]wire.Field
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
	if v.Check != nil {
		w, err = wire.NewValueI32(*(v.Check)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 2, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a NestedStruct struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a NestedStruct struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v NestedStruct
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *NestedStruct) FromWire(w wire.Value) error {
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
		case 2:
			if field.Value.Type() == wire.TI32 {
				var x int32
				x, err = field.Value.GetI32(), error(nil)
				v.Check = &x
				if err != nil {
					return err
				}

			}
		}
	}

	if !msgIsSet {
		return errors.New("field Msg of NestedStruct is required")
	}

	return nil
}

// String returns a readable string representation of a NestedStruct
// struct.
func (v *NestedStruct) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [2]string
	i := 0
	fields[i] = fmt.Sprintf("Msg: %v", v.Msg)
	i++
	if v.Check != nil {
		fields[i] = fmt.Sprintf("Check: %v", *(v.Check))
		i++
	}

	return fmt.Sprintf("NestedStruct{%v}", strings.Join(fields[:i], ", "))
}

func _I32_EqualsPtr(lhs, rhs *int32) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

// Equals returns true if all the fields of this NestedStruct match the
// provided NestedStruct.
//
// This function performs a deep comparison.
func (v *NestedStruct) Equals(rhs *NestedStruct) bool {
	if !(v.Msg == rhs.Msg) {
		return false
	}
	if !_I32_EqualsPtr(v.Check, rhs.Check) {
		return false
	}

	return true
}

// GetCheck returns the value of Check if it is set or its
// zero value if it is unset.
func (v *NestedStruct) GetCheck() (o int32) {
	if v.Check != nil {
		return *v.Check
	}

	return
}

type OtherAuthErr struct {
	Message string `json:"message,required"`
}

// ToWire translates a OtherAuthErr struct into a Thrift-level intermediate
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
func (v *OtherAuthErr) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueString(v.Message), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a OtherAuthErr struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a OtherAuthErr struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v OtherAuthErr
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *OtherAuthErr) FromWire(w wire.Value) error {
	var err error

	messageIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.Message, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				messageIsSet = true
			}
		}
	}

	if !messageIsSet {
		return errors.New("field Message of OtherAuthErr is required")
	}

	return nil
}

// String returns a readable string representation of a OtherAuthErr
// struct.
func (v *OtherAuthErr) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	fields[i] = fmt.Sprintf("Message: %v", v.Message)
	i++

	return fmt.Sprintf("OtherAuthErr{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this OtherAuthErr match the
// provided OtherAuthErr.
//
// This function performs a deep comparison.
func (v *OtherAuthErr) Equals(rhs *OtherAuthErr) bool {
	if !(v.Message == rhs.Message) {
		return false
	}

	return true
}

func (v *OtherAuthErr) Error() string {
	return v.String()
}

type ServerErr struct {
	Message string `json:"message,required"`
}

// ToWire translates a ServerErr struct into a Thrift-level intermediate
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
func (v *ServerErr) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueString(v.Message), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a ServerErr struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a ServerErr struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v ServerErr
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *ServerErr) FromWire(w wire.Value) error {
	var err error

	messageIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.Message, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				messageIsSet = true
			}
		}
	}

	if !messageIsSet {
		return errors.New("field Message of ServerErr is required")
	}

	return nil
}

// String returns a readable string representation of a ServerErr
// struct.
func (v *ServerErr) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	fields[i] = fmt.Sprintf("Message: %v", v.Message)
	i++

	return fmt.Sprintf("ServerErr{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this ServerErr match the
// provided ServerErr.
//
// This function performs a deep comparison.
func (v *ServerErr) Equals(rhs *ServerErr) bool {
	if !(v.Message == rhs.Message) {
		return false
	}

	return true
}

func (v *ServerErr) Error() string {
	return v.String()
}

type TransHeader struct {
}

// ToWire translates a TransHeader struct into a Thrift-level intermediate
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
func (v *TransHeader) ToWire() (wire.Value, error) {
	var (
		fields [0]wire.Field
		i      int = 0
	)

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a TransHeader struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a TransHeader struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v TransHeader
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *TransHeader) FromWire(w wire.Value) error {

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		}
	}

	return nil
}

// String returns a readable string representation of a TransHeader
// struct.
func (v *TransHeader) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [0]string
	i := 0

	return fmt.Sprintf("TransHeader{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this TransHeader match the
// provided TransHeader.
//
// This function performs a deep comparison.
func (v *TransHeader) Equals(rhs *TransHeader) bool {

	return true
}

type TransStruct struct {
	Message string        `json:"message,required"`
	Driver  *NestedStruct `json:"driver,omitempty"`
	Rider   *NestedStruct `json:"rider,required"`
}

// ToWire translates a TransStruct struct into a Thrift-level intermediate
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
func (v *TransStruct) ToWire() (wire.Value, error) {
	var (
		fields [3]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueString(v.Message), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++
	if v.Driver != nil {
		w, err = v.Driver.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 2, Value: w}
		i++
	}
	if v.Rider == nil {
		return w, errors.New("field Rider of TransStruct is required")
	}
	w, err = v.Rider.ToWire()
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 3, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _NestedStruct_Read(w wire.Value) (*NestedStruct, error) {
	var v NestedStruct
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a TransStruct struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a TransStruct struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v TransStruct
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *TransStruct) FromWire(w wire.Value) error {
	var err error

	messageIsSet := false

	riderIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.Message, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				messageIsSet = true
			}
		case 2:
			if field.Value.Type() == wire.TStruct {
				v.Driver, err = _NestedStruct_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 3:
			if field.Value.Type() == wire.TStruct {
				v.Rider, err = _NestedStruct_Read(field.Value)
				if err != nil {
					return err
				}
				riderIsSet = true
			}
		}
	}

	if !messageIsSet {
		return errors.New("field Message of TransStruct is required")
	}

	if !riderIsSet {
		return errors.New("field Rider of TransStruct is required")
	}

	return nil
}

// String returns a readable string representation of a TransStruct
// struct.
func (v *TransStruct) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [3]string
	i := 0
	fields[i] = fmt.Sprintf("Message: %v", v.Message)
	i++
	if v.Driver != nil {
		fields[i] = fmt.Sprintf("Driver: %v", v.Driver)
		i++
	}
	fields[i] = fmt.Sprintf("Rider: %v", v.Rider)
	i++

	return fmt.Sprintf("TransStruct{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this TransStruct match the
// provided TransStruct.
//
// This function performs a deep comparison.
func (v *TransStruct) Equals(rhs *TransStruct) bool {
	if !(v.Message == rhs.Message) {
		return false
	}
	if !((v.Driver == nil && rhs.Driver == nil) || (v.Driver != nil && rhs.Driver != nil && v.Driver.Equals(rhs.Driver))) {
		return false
	}
	if !v.Rider.Equals(rhs.Rider) {
		return false
	}

	return true
}

type UUID string

// ToWire translates UUID into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
func (v UUID) ToWire() (wire.Value, error) {
	x := (string)(v)
	return wire.NewValueString(x), error(nil)
}

// String returns a readable string representation of UUID.
func (v UUID) String() string {
	x := (string)(v)
	return fmt.Sprint(x)
}

// FromWire deserializes UUID from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
func (v *UUID) FromWire(w wire.Value) error {
	x, err := w.GetString(), error(nil)
	*v = (UUID)(x)
	return err
}

// Equals returns true if this UUID is equal to the provided
// UUID.
func (lhs UUID) Equals(rhs UUID) bool {
	return (lhs == rhs)
}
