// Code generated by zanzibar
// @generated
// Checksum : lChm9o0/jzl92mYbTSO6Ag==
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

func easyjson6d3473daDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceTransHeaders(in *jlexer.Lexer, out *SimpleService_TransHeaders_Result) {
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
					out.Success = new(TransHeader)
				}
				(*out.Success).UnmarshalEasyJSON(in)
			}
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
		case "otherAuthErr":
			if in.IsNull() {
				in.Skip()
				out.OtherAuthErr = nil
			} else {
				if out.OtherAuthErr == nil {
					out.OtherAuthErr = new(OtherAuthErr)
				}
				(*out.OtherAuthErr).UnmarshalEasyJSON(in)
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
func easyjson6d3473daEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceTransHeaders(out *jwriter.Writer, in SimpleService_TransHeaders_Result) {
	out.RawByte('{')
	first := true
	_ = first
	if in.Success != nil {
		const prefix string = ",\"success\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		(*in.Success).MarshalEasyJSON(out)
	}
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
	if in.OtherAuthErr != nil {
		const prefix string = ",\"otherAuthErr\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		(*in.OtherAuthErr).MarshalEasyJSON(out)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v SimpleService_TransHeaders_Result) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6d3473daEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceTransHeaders(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SimpleService_TransHeaders_Result) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6d3473daEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceTransHeaders(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *SimpleService_TransHeaders_Result) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6d3473daDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceTransHeaders(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SimpleService_TransHeaders_Result) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6d3473daDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceTransHeaders(l, v)
}
func easyjson6d3473daDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceTransHeaders1(in *jlexer.Lexer, out *SimpleService_TransHeaders_Args) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var ReqSet bool
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
		case "req":
			if in.IsNull() {
				in.Skip()
				out.Req = nil
			} else {
				if out.Req == nil {
					out.Req = new(TransHeader)
				}
				(*out.Req).UnmarshalEasyJSON(in)
			}
			ReqSet = true
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !ReqSet {
		in.AddError(fmt.Errorf("key 'req' is required"))
	}
}
func easyjson6d3473daEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceTransHeaders1(out *jwriter.Writer, in SimpleService_TransHeaders_Args) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"req\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Req == nil {
			out.RawString("null")
		} else {
			(*in.Req).MarshalEasyJSON(out)
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v SimpleService_TransHeaders_Args) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6d3473daEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceTransHeaders1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SimpleService_TransHeaders_Args) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6d3473daEncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceTransHeaders1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *SimpleService_TransHeaders_Args) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6d3473daDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceTransHeaders1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SimpleService_TransHeaders_Args) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6d3473daDecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBazBazSimpleServiceTransHeaders1(l, v)
}
