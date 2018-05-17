// Code generated by zanzibar
// @generated
// Checksum : g8OOOcCso4XNvQ3nErQeOg==
// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package bar

import (
	json "encoding/json"
	fmt "fmt"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar(in *jlexer.Lexer, out *QueryParamsStruct) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var NameSet bool
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "name":
			out.Name = string(in.String())
			NameSet = true
		case "userUUID":
			if in.IsNull() {
				in.Skip()
				out.UserUUID = nil
			} else {
				if out.UserUUID == nil {
					out.UserUUID = new(string)
				}
				*out.UserUUID = string(in.String())
			}
		case "authUUID":
			if in.IsNull() {
				in.Skip()
				out.AuthUUID = nil
			} else {
				if out.AuthUUID == nil {
					out.AuthUUID = new(string)
				}
				*out.AuthUUID = string(in.String())
			}
		case "authUUID2":
			if in.IsNull() {
				in.Skip()
				out.AuthUUID2 = nil
			} else {
				if out.AuthUUID2 == nil {
					out.AuthUUID2 = new(string)
				}
				*out.AuthUUID2 = string(in.String())
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !NameSet {
		in.AddError(fmt.Errorf("key 'name' is required"))
	}
}
func easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar(out *jwriter.Writer, in QueryParamsStruct) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"name\":")
	out.String(string(in.Name))
	if in.UserUUID != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"userUUID\":")
		if in.UserUUID == nil {
			out.RawString("null")
		} else {
			out.String(string(*in.UserUUID))
		}
	}
	if in.AuthUUID != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"authUUID\":")
		if in.AuthUUID == nil {
			out.RawString("null")
		} else {
			out.String(string(*in.AuthUUID))
		}
	}
	if in.AuthUUID2 != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"authUUID2\":")
		if in.AuthUUID2 == nil {
			out.RawString("null")
		} else {
			out.String(string(*in.AuthUUID2))
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v QueryParamsStruct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v QueryParamsStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *QueryParamsStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *QueryParamsStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar(l, v)
}
func easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar1(in *jlexer.Lexer, out *QueryParamsOptsStruct) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var NameSet bool
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "name":
			out.Name = string(in.String())
			NameSet = true
		case "userUUID":
			if in.IsNull() {
				in.Skip()
				out.UserUUID = nil
			} else {
				if out.UserUUID == nil {
					out.UserUUID = new(string)
				}
				*out.UserUUID = string(in.String())
			}
		case "authUUID":
			if in.IsNull() {
				in.Skip()
				out.AuthUUID = nil
			} else {
				if out.AuthUUID == nil {
					out.AuthUUID = new(string)
				}
				*out.AuthUUID = string(in.String())
			}
		case "authUUID2":
			if in.IsNull() {
				in.Skip()
				out.AuthUUID2 = nil
			} else {
				if out.AuthUUID2 == nil {
					out.AuthUUID2 = new(string)
				}
				*out.AuthUUID2 = string(in.String())
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !NameSet {
		in.AddError(fmt.Errorf("key 'name' is required"))
	}
}
func easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar1(out *jwriter.Writer, in QueryParamsOptsStruct) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"name\":")
	out.String(string(in.Name))
	if in.UserUUID != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"userUUID\":")
		if in.UserUUID == nil {
			out.RawString("null")
		} else {
			out.String(string(*in.UserUUID))
		}
	}
	if in.AuthUUID != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"authUUID\":")
		if in.AuthUUID == nil {
			out.RawString("null")
		} else {
			out.String(string(*in.AuthUUID))
		}
	}
	if in.AuthUUID2 != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"authUUID2\":")
		if in.AuthUUID2 == nil {
			out.RawString("null")
		} else {
			out.String(string(*in.AuthUUID2))
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v QueryParamsOptsStruct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v QueryParamsOptsStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *QueryParamsOptsStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *QueryParamsOptsStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar1(l, v)
}
func easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar2(in *jlexer.Lexer, out *ParamsStruct) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar2(out *jwriter.Writer, in ParamsStruct) {
	out.RawByte('{')
	first := true
	_ = first
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ParamsStruct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ParamsStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ParamsStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ParamsStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar2(l, v)
}
func easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar3(in *jlexer.Lexer, out *BarResponse) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var StringFieldSet bool
	var IntWithRangeSet bool
	var IntWithoutRangeSet bool
	var MapIntWithRangeSet bool
	var MapIntWithoutRangeSet bool
	var BinaryFieldSet bool
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "stringField":
			out.StringField = string(in.String())
			StringFieldSet = true
		case "intWithRange":
			out.IntWithRange = int32(in.Int32())
			IntWithRangeSet = true
		case "intWithoutRange":
			out.IntWithoutRange = int32(in.Int32())
			IntWithoutRangeSet = true
		case "mapIntWithRange":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.MapIntWithRange = make(map[UUID]int32)
				} else {
					out.MapIntWithRange = nil
				}
				for !in.IsDelim('}') {
					key := UUID(in.String())
					in.WantColon()
					var v1 int32
					v1 = int32(in.Int32())
					(out.MapIntWithRange)[key] = v1
					in.WantComma()
				}
				in.Delim('}')
			}
			MapIntWithRangeSet = true
		case "mapIntWithoutRange":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.MapIntWithoutRange = make(map[string]int32)
				} else {
					out.MapIntWithoutRange = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v2 int32
					v2 = int32(in.Int32())
					(out.MapIntWithoutRange)[key] = v2
					in.WantComma()
				}
				in.Delim('}')
			}
			MapIntWithoutRangeSet = true
		case "binaryField":
			if in.IsNull() {
				in.Skip()
				out.BinaryField = nil
			} else {
				out.BinaryField = in.Bytes()
			}
			BinaryFieldSet = true
		case "nextResponse":
			if in.IsNull() {
				in.Skip()
				out.NextResponse = nil
			} else {
				if out.NextResponse == nil {
					out.NextResponse = new(BarResponse)
				}
				(*out.NextResponse).UnmarshalEasyJSON(in)
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !StringFieldSet {
		in.AddError(fmt.Errorf("key 'stringField' is required"))
	}
	if !IntWithRangeSet {
		in.AddError(fmt.Errorf("key 'intWithRange' is required"))
	}
	if !IntWithoutRangeSet {
		in.AddError(fmt.Errorf("key 'intWithoutRange' is required"))
	}
	if !MapIntWithRangeSet {
		in.AddError(fmt.Errorf("key 'mapIntWithRange' is required"))
	}
	if !MapIntWithoutRangeSet {
		in.AddError(fmt.Errorf("key 'mapIntWithoutRange' is required"))
	}
	if !BinaryFieldSet {
		in.AddError(fmt.Errorf("key 'binaryField' is required"))
	}
}
func easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar3(out *jwriter.Writer, in BarResponse) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"stringField\":")
	out.String(string(in.StringField))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"intWithRange\":")
	out.Int32(int32(in.IntWithRange))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"intWithoutRange\":")
	out.Int32(int32(in.IntWithoutRange))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"mapIntWithRange\":")
	if in.MapIntWithRange == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
		out.RawString(`null`)
	} else {
		out.RawByte('{')
		v4First := true
		for v4Name, v4Value := range in.MapIntWithRange {
			if !v4First {
				out.RawByte(',')
			}
			v4First = false
			out.String(string(v4Name))
			out.RawByte(':')
			out.Int32(int32(v4Value))
		}
		out.RawByte('}')
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"mapIntWithoutRange\":")
	if in.MapIntWithoutRange == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
		out.RawString(`null`)
	} else {
		out.RawByte('{')
		v5First := true
		for v5Name, v5Value := range in.MapIntWithoutRange {
			if !v5First {
				out.RawByte(',')
			}
			v5First = false
			out.String(string(v5Name))
			out.RawByte(':')
			out.Int32(int32(v5Value))
		}
		out.RawByte('}')
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"binaryField\":")
	out.Base64Bytes(in.BinaryField)
	if in.NextResponse != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"nextResponse\":")
		if in.NextResponse == nil {
			out.RawString("null")
		} else {
			(*in.NextResponse).MarshalEasyJSON(out)
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v BarResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v BarResponse) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *BarResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *BarResponse) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar3(l, v)
}
func easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar4(in *jlexer.Lexer, out *BarRequest) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var BoolFieldSet bool
	var BinaryFieldSet bool
	var TimestampSet bool
	var EnumFieldSet bool
	var LongFieldSet bool
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "stringField":
			if in.IsNull() {
				in.Skip()
				out.StringField = nil
			} else {
				if out.StringField == nil {
					out.StringField = new(string)
				}
				*out.StringField = string(in.String())
			}
		case "boolField":
			out.BoolField = bool(in.Bool())
			BoolFieldSet = true
		case "binaryField":
			if in.IsNull() {
				in.Skip()
				out.BinaryField = nil
			} else {
				out.BinaryField = in.Bytes()
			}
			BinaryFieldSet = true
		case "timestamp":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Timestamp).UnmarshalJSON(data))
			}
			TimestampSet = true
		case "enumField":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.EnumField).UnmarshalJSON(data))
			}
			EnumFieldSet = true
		case "longField":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.LongField).UnmarshalJSON(data))
			}
			LongFieldSet = true
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !BoolFieldSet {
		in.AddError(fmt.Errorf("key 'boolField' is required"))
	}
	if !BinaryFieldSet {
		in.AddError(fmt.Errorf("key 'binaryField' is required"))
	}
	if !TimestampSet {
		in.AddError(fmt.Errorf("key 'timestamp' is required"))
	}
	if !EnumFieldSet {
		in.AddError(fmt.Errorf("key 'enumField' is required"))
	}
	if !LongFieldSet {
		in.AddError(fmt.Errorf("key 'longField' is required"))
	}
}
func easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar4(out *jwriter.Writer, in BarRequest) {
	out.RawByte('{')
	first := true
	_ = first
	if in.StringField != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"stringField\":")
		if in.StringField == nil {
			out.RawString("null")
		} else {
			out.String(string(*in.StringField))
		}
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"boolField\":")
	out.Bool(bool(in.BoolField))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"binaryField\":")
	out.Base64Bytes(in.BinaryField)
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"timestamp\":")
	out.Raw((in.Timestamp).MarshalJSON())
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"enumField\":")
	out.Raw((in.EnumField).MarshalJSON())
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"longField\":")
	out.Raw((in.LongField).MarshalJSON())
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v BarRequest) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar4(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v BarRequest) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar4(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *BarRequest) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar4(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *BarRequest) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar4(l, v)
}
func easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar5(in *jlexer.Lexer, out *BarException) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var StringFieldSet bool
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "stringField":
			out.StringField = string(in.String())
			StringFieldSet = true
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !StringFieldSet {
		in.AddError(fmt.Errorf("key 'stringField' is required"))
	}
}
func easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar5(out *jwriter.Writer, in BarException) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"stringField\":")
	out.String(string(in.StringField))
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v BarException) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar5(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v BarException) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6601e8cdEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar5(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *BarException) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar5(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *BarException) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6601e8cdDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar5(l, v)
}
