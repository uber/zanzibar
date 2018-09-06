// Code generated by zanzibar
// @generated
// Checksum : zfzgihqPmbjd7iX5b79+jg==
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

func easyjson8b993ec5DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCallPanic(in *jlexer.Lexer, out *SimpleService_CallPanic_Result) {
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
		case "authErr":
			if in.IsNull() {
				in.Skip()
				out.AuthErr = nil
			} else {
				if out.AuthErr == nil {
					out.AuthErr = new(AuthErr)
				}
				(*out.AuthErr).UnmarshalEasyJSON(in)
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
func easyjson8b993ec5EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCallPanic(out *jwriter.Writer, in SimpleService_CallPanic_Result) {
	out.RawByte('{')
	first := true
	_ = first
	if in.AuthErr != nil {
		const prefix string = ",\"authErr\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		(*in.AuthErr).MarshalEasyJSON(out)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v SimpleService_CallPanic_Result) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson8b993ec5EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCallPanic(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SimpleService_CallPanic_Result) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson8b993ec5EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCallPanic(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *SimpleService_CallPanic_Result) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson8b993ec5DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCallPanic(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SimpleService_CallPanic_Result) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson8b993ec5DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCallPanic(l, v)
}
func easyjson8b993ec5DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCallPanic1(in *jlexer.Lexer, out *SimpleService_CallPanic_Args) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var ArgSet bool
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
		case "arg":
			if in.IsNull() {
				in.Skip()
				out.Arg = nil
			} else {
				if out.Arg == nil {
					out.Arg = new(BazRequest)
				}
				(*out.Arg).UnmarshalEasyJSON(in)
			}
			ArgSet = true
		case "i64Optional":
			if in.IsNull() {
				in.Skip()
				out.I64Optional = nil
			} else {
				if out.I64Optional == nil {
					out.I64Optional = new(int64)
				}
				*out.I64Optional = int64(in.Int64())
			}
		case "testUUID":
			if in.IsNull() {
				in.Skip()
				out.TestUUID = nil
			} else {
				if out.TestUUID == nil {
					out.TestUUID = new(UUID)
				}
				*out.TestUUID = UUID(in.String())
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
	if !ArgSet {
		in.AddError(fmt.Errorf("key 'arg' is required"))
	}
}
func easyjson8b993ec5EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCallPanic1(out *jwriter.Writer, in SimpleService_CallPanic_Args) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"arg\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Arg == nil {
			out.RawString("null")
		} else {
			(*in.Arg).MarshalEasyJSON(out)
		}
	}
	if in.I64Optional != nil {
		const prefix string = ",\"i64Optional\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(*in.I64Optional))
	}
	if in.TestUUID != nil {
		const prefix string = ",\"testUUID\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(*in.TestUUID))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v SimpleService_CallPanic_Args) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson8b993ec5EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCallPanic1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SimpleService_CallPanic_Args) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson8b993ec5EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCallPanic1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *SimpleService_CallPanic_Args) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson8b993ec5DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCallPanic1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SimpleService_CallPanic_Args) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson8b993ec5DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceCallPanic1(l, v)
}