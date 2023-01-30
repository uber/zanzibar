// Code generated by thriftrw v1.29.2. DO NOT EDIT.
// @generated

package googlenow

import (
	errors "errors"
	fmt "fmt"
	stream "go.uber.org/thriftrw/protocol/stream"
	wire "go.uber.org/thriftrw/wire"
	zapcore "go.uber.org/zap/zapcore"
	strings "strings"
)

// GoogleNowService_AddCredentials_Args represents the arguments for the GoogleNowService.addCredentials function.
//
// The arguments for addCredentials are sent and received over the wire as this struct.
type GoogleNowService_AddCredentials_Args struct {
	AuthCode string `json:"authCode,required"`
}

// ToWire translates a GoogleNowService_AddCredentials_Args struct into a Thrift-level intermediate
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
func (v *GoogleNowService_AddCredentials_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueString(v.AuthCode), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a GoogleNowService_AddCredentials_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a GoogleNowService_AddCredentials_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v GoogleNowService_AddCredentials_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *GoogleNowService_AddCredentials_Args) FromWire(w wire.Value) error {
	var err error

	authCodeIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.AuthCode, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				authCodeIsSet = true
			}
		}
	}

	if !authCodeIsSet {
		return errors.New("field AuthCode of GoogleNowService_AddCredentials_Args is required")
	}

	return nil
}

// Encode serializes a GoogleNowService_AddCredentials_Args struct directly into bytes, without going
// through an intermediary type.
//
// An error is returned if a GoogleNowService_AddCredentials_Args struct could not be encoded.
func (v *GoogleNowService_AddCredentials_Args) Encode(sw stream.Writer) error {
	if err := sw.WriteStructBegin(); err != nil {
		return err
	}

	if err := sw.WriteFieldBegin(stream.FieldHeader{ID: 1, Type: wire.TBinary}); err != nil {
		return err
	}
	if err := sw.WriteString(v.AuthCode); err != nil {
		return err
	}
	if err := sw.WriteFieldEnd(); err != nil {
		return err
	}

	return sw.WriteStructEnd()
}

// Decode deserializes a GoogleNowService_AddCredentials_Args struct directly from its Thrift-level
// representation, without going through an intemediary type.
//
// An error is returned if a GoogleNowService_AddCredentials_Args struct could not be generated from the wire
// representation.
func (v *GoogleNowService_AddCredentials_Args) Decode(sr stream.Reader) error {

	authCodeIsSet := false

	if err := sr.ReadStructBegin(); err != nil {
		return err
	}

	fh, ok, err := sr.ReadFieldBegin()
	if err != nil {
		return err
	}

	for ok {
		switch {
		case fh.ID == 1 && fh.Type == wire.TBinary:
			v.AuthCode, err = sr.ReadString()
			if err != nil {
				return err
			}
			authCodeIsSet = true
		default:
			if err := sr.Skip(fh.Type); err != nil {
				return err
			}
		}

		if err := sr.ReadFieldEnd(); err != nil {
			return err
		}

		if fh, ok, err = sr.ReadFieldBegin(); err != nil {
			return err
		}
	}

	if err := sr.ReadStructEnd(); err != nil {
		return err
	}

	if !authCodeIsSet {
		return errors.New("field AuthCode of GoogleNowService_AddCredentials_Args is required")
	}

	return nil
}

// String returns a readable string representation of a GoogleNowService_AddCredentials_Args
// struct.
func (v *GoogleNowService_AddCredentials_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	fields[i] = fmt.Sprintf("AuthCode: %v", v.AuthCode)
	i++

	return fmt.Sprintf("GoogleNowService_AddCredentials_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this GoogleNowService_AddCredentials_Args match the
// provided GoogleNowService_AddCredentials_Args.
//
// This function performs a deep comparison.
func (v *GoogleNowService_AddCredentials_Args) Equals(rhs *GoogleNowService_AddCredentials_Args) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !(v.AuthCode == rhs.AuthCode) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of GoogleNowService_AddCredentials_Args.
func (v *GoogleNowService_AddCredentials_Args) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	enc.AddString("authCode", v.AuthCode)
	return err
}

// GetAuthCode returns the value of AuthCode if it is set or its
// zero value if it is unset.
func (v *GoogleNowService_AddCredentials_Args) GetAuthCode() (o string) {
	if v != nil {
		o = v.AuthCode
	}
	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "addCredentials" for this struct.
func (v *GoogleNowService_AddCredentials_Args) MethodName() string {
	return "addCredentials"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *GoogleNowService_AddCredentials_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// GoogleNowService_AddCredentials_Helper provides functions that aid in handling the
// parameters and return values of the GoogleNowService.addCredentials
// function.
var GoogleNowService_AddCredentials_Helper = struct {
	// Args accepts the parameters of addCredentials in-order and returns
	// the arguments struct for the function.
	Args func(
		authCode string,
	) *GoogleNowService_AddCredentials_Args

	// IsException returns true if the given error can be thrown
	// by addCredentials.
	//
	// An error can be thrown by addCredentials only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for addCredentials
	// given the error returned by it. The provided error may
	// be nil if addCredentials did not fail.
	//
	// This allows mapping errors returned by addCredentials into a
	// serializable result struct. WrapResponse returns a
	// non-nil error if the provided error cannot be thrown by
	// addCredentials
	//
	//   err := addCredentials(args)
	//   result, err := GoogleNowService_AddCredentials_Helper.WrapResponse(err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from addCredentials: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(error) (*GoogleNowService_AddCredentials_Result, error)

	// UnwrapResponse takes the result struct for addCredentials
	// and returns the erorr returned by it (if any).
	//
	// The error is non-nil only if addCredentials threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   err := GoogleNowService_AddCredentials_Helper.UnwrapResponse(result)
	UnwrapResponse func(*GoogleNowService_AddCredentials_Result) error
}{}

func init() {
	GoogleNowService_AddCredentials_Helper.Args = func(
		authCode string,
	) *GoogleNowService_AddCredentials_Args {
		return &GoogleNowService_AddCredentials_Args{
			AuthCode: authCode,
		}
	}

	GoogleNowService_AddCredentials_Helper.IsException = func(err error) bool {
		switch err.(type) {
		default:
			return false
		}
	}

	GoogleNowService_AddCredentials_Helper.WrapResponse = func(err error) (*GoogleNowService_AddCredentials_Result, error) {
		if err == nil {
			return &GoogleNowService_AddCredentials_Result{}, nil
		}

		return nil, err
	}
	GoogleNowService_AddCredentials_Helper.UnwrapResponse = func(result *GoogleNowService_AddCredentials_Result) (err error) {
		return
	}

}

// GoogleNowService_AddCredentials_Result represents the result of a GoogleNowService.addCredentials function call.
//
// The result of a addCredentials execution is sent and received over the wire as this struct.
type GoogleNowService_AddCredentials_Result struct {
}

// ToWire translates a GoogleNowService_AddCredentials_Result struct into a Thrift-level intermediate
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
func (v *GoogleNowService_AddCredentials_Result) ToWire() (wire.Value, error) {
	var (
		fields [0]wire.Field
		i      int = 0
	)

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a GoogleNowService_AddCredentials_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a GoogleNowService_AddCredentials_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v GoogleNowService_AddCredentials_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *GoogleNowService_AddCredentials_Result) FromWire(w wire.Value) error {

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		}
	}

	return nil
}

// Encode serializes a GoogleNowService_AddCredentials_Result struct directly into bytes, without going
// through an intermediary type.
//
// An error is returned if a GoogleNowService_AddCredentials_Result struct could not be encoded.
func (v *GoogleNowService_AddCredentials_Result) Encode(sw stream.Writer) error {
	if err := sw.WriteStructBegin(); err != nil {
		return err
	}

	return sw.WriteStructEnd()
}

// Decode deserializes a GoogleNowService_AddCredentials_Result struct directly from its Thrift-level
// representation, without going through an intemediary type.
//
// An error is returned if a GoogleNowService_AddCredentials_Result struct could not be generated from the wire
// representation.
func (v *GoogleNowService_AddCredentials_Result) Decode(sr stream.Reader) error {

	if err := sr.ReadStructBegin(); err != nil {
		return err
	}

	fh, ok, err := sr.ReadFieldBegin()
	if err != nil {
		return err
	}

	for ok {
		switch {
		default:
			if err := sr.Skip(fh.Type); err != nil {
				return err
			}
		}

		if err := sr.ReadFieldEnd(); err != nil {
			return err
		}

		if fh, ok, err = sr.ReadFieldBegin(); err != nil {
			return err
		}
	}

	if err := sr.ReadStructEnd(); err != nil {
		return err
	}

	return nil
}

// String returns a readable string representation of a GoogleNowService_AddCredentials_Result
// struct.
func (v *GoogleNowService_AddCredentials_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [0]string
	i := 0

	return fmt.Sprintf("GoogleNowService_AddCredentials_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this GoogleNowService_AddCredentials_Result match the
// provided GoogleNowService_AddCredentials_Result.
//
// This function performs a deep comparison.
func (v *GoogleNowService_AddCredentials_Result) Equals(rhs *GoogleNowService_AddCredentials_Result) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of GoogleNowService_AddCredentials_Result.
func (v *GoogleNowService_AddCredentials_Result) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	return err
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "addCredentials" for this struct.
func (v *GoogleNowService_AddCredentials_Result) MethodName() string {
	return "addCredentials"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *GoogleNowService_AddCredentials_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}

// GoogleNowService_CheckCredentials_Args represents the arguments for the GoogleNowService.checkCredentials function.
//
// The arguments for checkCredentials are sent and received over the wire as this struct.
type GoogleNowService_CheckCredentials_Args struct {
}

// ToWire translates a GoogleNowService_CheckCredentials_Args struct into a Thrift-level intermediate
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
func (v *GoogleNowService_CheckCredentials_Args) ToWire() (wire.Value, error) {
	var (
		fields [0]wire.Field
		i      int = 0
	)

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a GoogleNowService_CheckCredentials_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a GoogleNowService_CheckCredentials_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v GoogleNowService_CheckCredentials_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *GoogleNowService_CheckCredentials_Args) FromWire(w wire.Value) error {

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		}
	}

	return nil
}

// Encode serializes a GoogleNowService_CheckCredentials_Args struct directly into bytes, without going
// through an intermediary type.
//
// An error is returned if a GoogleNowService_CheckCredentials_Args struct could not be encoded.
func (v *GoogleNowService_CheckCredentials_Args) Encode(sw stream.Writer) error {
	if err := sw.WriteStructBegin(); err != nil {
		return err
	}

	return sw.WriteStructEnd()
}

// Decode deserializes a GoogleNowService_CheckCredentials_Args struct directly from its Thrift-level
// representation, without going through an intemediary type.
//
// An error is returned if a GoogleNowService_CheckCredentials_Args struct could not be generated from the wire
// representation.
func (v *GoogleNowService_CheckCredentials_Args) Decode(sr stream.Reader) error {

	if err := sr.ReadStructBegin(); err != nil {
		return err
	}

	fh, ok, err := sr.ReadFieldBegin()
	if err != nil {
		return err
	}

	for ok {
		switch {
		default:
			if err := sr.Skip(fh.Type); err != nil {
				return err
			}
		}

		if err := sr.ReadFieldEnd(); err != nil {
			return err
		}

		if fh, ok, err = sr.ReadFieldBegin(); err != nil {
			return err
		}
	}

	if err := sr.ReadStructEnd(); err != nil {
		return err
	}

	return nil
}

// String returns a readable string representation of a GoogleNowService_CheckCredentials_Args
// struct.
func (v *GoogleNowService_CheckCredentials_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [0]string
	i := 0

	return fmt.Sprintf("GoogleNowService_CheckCredentials_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this GoogleNowService_CheckCredentials_Args match the
// provided GoogleNowService_CheckCredentials_Args.
//
// This function performs a deep comparison.
func (v *GoogleNowService_CheckCredentials_Args) Equals(rhs *GoogleNowService_CheckCredentials_Args) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of GoogleNowService_CheckCredentials_Args.
func (v *GoogleNowService_CheckCredentials_Args) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	return err
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "checkCredentials" for this struct.
func (v *GoogleNowService_CheckCredentials_Args) MethodName() string {
	return "checkCredentials"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *GoogleNowService_CheckCredentials_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// GoogleNowService_CheckCredentials_Helper provides functions that aid in handling the
// parameters and return values of the GoogleNowService.checkCredentials
// function.
var GoogleNowService_CheckCredentials_Helper = struct {
	// Args accepts the parameters of checkCredentials in-order and returns
	// the arguments struct for the function.
	Args func() *GoogleNowService_CheckCredentials_Args

	// IsException returns true if the given error can be thrown
	// by checkCredentials.
	//
	// An error can be thrown by checkCredentials only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for checkCredentials
	// given the error returned by it. The provided error may
	// be nil if checkCredentials did not fail.
	//
	// This allows mapping errors returned by checkCredentials into a
	// serializable result struct. WrapResponse returns a
	// non-nil error if the provided error cannot be thrown by
	// checkCredentials
	//
	//   err := checkCredentials(args)
	//   result, err := GoogleNowService_CheckCredentials_Helper.WrapResponse(err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from checkCredentials: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(error) (*GoogleNowService_CheckCredentials_Result, error)

	// UnwrapResponse takes the result struct for checkCredentials
	// and returns the erorr returned by it (if any).
	//
	// The error is non-nil only if checkCredentials threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   err := GoogleNowService_CheckCredentials_Helper.UnwrapResponse(result)
	UnwrapResponse func(*GoogleNowService_CheckCredentials_Result) error
}{}

func init() {
	GoogleNowService_CheckCredentials_Helper.Args = func() *GoogleNowService_CheckCredentials_Args {
		return &GoogleNowService_CheckCredentials_Args{}
	}

	GoogleNowService_CheckCredentials_Helper.IsException = func(err error) bool {
		switch err.(type) {
		default:
			return false
		}
	}

	GoogleNowService_CheckCredentials_Helper.WrapResponse = func(err error) (*GoogleNowService_CheckCredentials_Result, error) {
		if err == nil {
			return &GoogleNowService_CheckCredentials_Result{}, nil
		}

		return nil, err
	}
	GoogleNowService_CheckCredentials_Helper.UnwrapResponse = func(result *GoogleNowService_CheckCredentials_Result) (err error) {
		return
	}

}

// GoogleNowService_CheckCredentials_Result represents the result of a GoogleNowService.checkCredentials function call.
//
// The result of a checkCredentials execution is sent and received over the wire as this struct.
type GoogleNowService_CheckCredentials_Result struct {
}

// ToWire translates a GoogleNowService_CheckCredentials_Result struct into a Thrift-level intermediate
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
func (v *GoogleNowService_CheckCredentials_Result) ToWire() (wire.Value, error) {
	var (
		fields [0]wire.Field
		i      int = 0
	)

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a GoogleNowService_CheckCredentials_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a GoogleNowService_CheckCredentials_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v GoogleNowService_CheckCredentials_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *GoogleNowService_CheckCredentials_Result) FromWire(w wire.Value) error {

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		}
	}

	return nil
}

// Encode serializes a GoogleNowService_CheckCredentials_Result struct directly into bytes, without going
// through an intermediary type.
//
// An error is returned if a GoogleNowService_CheckCredentials_Result struct could not be encoded.
func (v *GoogleNowService_CheckCredentials_Result) Encode(sw stream.Writer) error {
	if err := sw.WriteStructBegin(); err != nil {
		return err
	}

	return sw.WriteStructEnd()
}

// Decode deserializes a GoogleNowService_CheckCredentials_Result struct directly from its Thrift-level
// representation, without going through an intemediary type.
//
// An error is returned if a GoogleNowService_CheckCredentials_Result struct could not be generated from the wire
// representation.
func (v *GoogleNowService_CheckCredentials_Result) Decode(sr stream.Reader) error {

	if err := sr.ReadStructBegin(); err != nil {
		return err
	}

	fh, ok, err := sr.ReadFieldBegin()
	if err != nil {
		return err
	}

	for ok {
		switch {
		default:
			if err := sr.Skip(fh.Type); err != nil {
				return err
			}
		}

		if err := sr.ReadFieldEnd(); err != nil {
			return err
		}

		if fh, ok, err = sr.ReadFieldBegin(); err != nil {
			return err
		}
	}

	if err := sr.ReadStructEnd(); err != nil {
		return err
	}

	return nil
}

// String returns a readable string representation of a GoogleNowService_CheckCredentials_Result
// struct.
func (v *GoogleNowService_CheckCredentials_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [0]string
	i := 0

	return fmt.Sprintf("GoogleNowService_CheckCredentials_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this GoogleNowService_CheckCredentials_Result match the
// provided GoogleNowService_CheckCredentials_Result.
//
// This function performs a deep comparison.
func (v *GoogleNowService_CheckCredentials_Result) Equals(rhs *GoogleNowService_CheckCredentials_Result) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of GoogleNowService_CheckCredentials_Result.
func (v *GoogleNowService_CheckCredentials_Result) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	return err
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "checkCredentials" for this struct.
func (v *GoogleNowService_CheckCredentials_Result) MethodName() string {
	return "checkCredentials"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *GoogleNowService_CheckCredentials_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
