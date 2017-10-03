// Code generated by zanzibar
// @generated
// Checksum : S6XFoEeTSzz/tlj0qfW30w==
// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package baz

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

func easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCompare(in *jlexer.Lexer, out *SimpleService_Compare_Result) {
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
		case "success":
			if in.IsNull() {
				in.Skip()
				out.Success = nil
			} else {
				if out.Success == nil {
					out.Success = new(BazResponse)
				}
				easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz(in, &*out.Success)
			}
		case "authErr":
			if in.IsNull() {
				in.Skip()
				out.AuthErr = nil
			} else {
				if out.AuthErr == nil {
					out.AuthErr = new(AuthErr)
				}
				easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz1(in, &*out.AuthErr)
			}
		case "otherAuthErr":
			if in.IsNull() {
				in.Skip()
				out.OtherAuthErr = nil
			} else {
				if out.OtherAuthErr == nil {
					out.OtherAuthErr = new(OtherAuthErr)
				}
				easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz2(in, &*out.OtherAuthErr)
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
}
func easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCompare(out *jwriter.Writer, in SimpleService_Compare_Result) {
	out.RawByte('{')
	first := true
	_ = first
	if in.Success != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"success\":")
		if in.Success == nil {
			out.RawString("null")
		} else {
			easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz(out, *in.Success)
		}
	}
	if in.AuthErr != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"authErr\":")
		if in.AuthErr == nil {
			out.RawString("null")
		} else {
			easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz1(out, *in.AuthErr)
		}
	}
	if in.OtherAuthErr != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"otherAuthErr\":")
		if in.OtherAuthErr == nil {
			out.RawString("null")
		} else {
			easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz2(out, *in.OtherAuthErr)
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v SimpleService_Compare_Result) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCompare(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SimpleService_Compare_Result) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCompare(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *SimpleService_Compare_Result) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCompare(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SimpleService_Compare_Result) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCompare(l, v)
}
func easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz2(in *jlexer.Lexer, out *OtherAuthErr) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var MessageSet bool
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
		case "message":
			out.Message = string(in.String())
			MessageSet = true
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !MessageSet {
		in.AddError(fmt.Errorf("key 'message' is required"))
	}
}
func easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz2(out *jwriter.Writer, in OtherAuthErr) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"message\":")
	out.String(string(in.Message))
	out.RawByte('}')
}
func easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz1(in *jlexer.Lexer, out *AuthErr) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var MessageSet bool
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
		case "message":
			out.Message = string(in.String())
			MessageSet = true
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !MessageSet {
		in.AddError(fmt.Errorf("key 'message' is required"))
	}
}
func easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz1(out *jwriter.Writer, in AuthErr) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"message\":")
	out.String(string(in.Message))
	out.RawByte('}')
}
func easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz(in *jlexer.Lexer, out *BazResponse) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var MessageSet bool
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
		case "message":
			out.Message = string(in.String())
			MessageSet = true
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !MessageSet {
		in.AddError(fmt.Errorf("key 'message' is required"))
	}
}
func easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz(out *jwriter.Writer, in BazResponse) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"message\":")
	out.String(string(in.Message))
	out.RawByte('}')
}
func easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCompare1(in *jlexer.Lexer, out *SimpleService_Compare_Args) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var Arg1Set bool
	var Arg2Set bool
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
		case "arg1":
			if in.IsNull() {
				in.Skip()
				out.Arg1 = nil
			} else {
				if out.Arg1 == nil {
					out.Arg1 = new(BazRequest)
				}
				easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz3(in, &*out.Arg1)
			}
			Arg1Set = true
		case "arg2":
			if in.IsNull() {
				in.Skip()
				out.Arg2 = nil
			} else {
				if out.Arg2 == nil {
					out.Arg2 = new(BazRequest)
				}
				easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz3(in, &*out.Arg2)
			}
			Arg2Set = true
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !Arg1Set {
		in.AddError(fmt.Errorf("key 'arg1' is required"))
	}
	if !Arg2Set {
		in.AddError(fmt.Errorf("key 'arg2' is required"))
	}
}
func easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCompare1(out *jwriter.Writer, in SimpleService_Compare_Args) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"arg1\":")
	if in.Arg1 == nil {
		out.RawString("null")
	} else {
		easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz3(out, *in.Arg1)
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"arg2\":")
	if in.Arg2 == nil {
		out.RawString("null")
	} else {
		easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz3(out, *in.Arg2)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v SimpleService_Compare_Args) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCompare1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SimpleService_Compare_Args) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCompare1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *SimpleService_Compare_Args) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCompare1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SimpleService_Compare_Args) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCompare1(l, v)
}
func easyjson4d8a931DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz3(in *jlexer.Lexer, out *BazRequest) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var B1Set bool
	var S2Set bool
	var I3Set bool
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
		case "b1":
			out.B1 = bool(in.Bool())
			B1Set = true
		case "s2":
			out.S2 = string(in.String())
			S2Set = true
		case "i3":
			out.I3 = int32(in.Int32())
			I3Set = true
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !B1Set {
		in.AddError(fmt.Errorf("key 'b1' is required"))
	}
	if !S2Set {
		in.AddError(fmt.Errorf("key 's2' is required"))
	}
	if !I3Set {
		in.AddError(fmt.Errorf("key 'i3' is required"))
	}
}
func easyjson4d8a931EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBaz3(out *jwriter.Writer, in BazRequest) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"b1\":")
	out.Bool(bool(in.B1))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"s2\":")
	out.String(string(in.S2))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"i3\":")
	out.Int32(int32(in.I3))
	out.RawByte('}')
}
