// Code generated by zanzibar
// @generated
// Checksum : DyAItRSesStN0w++Bz/DZw==
// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package multi

import (
	json "encoding/json"
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

func easyjsonAfe36c91DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsMultiMultiServiceBFrontHello(in *jlexer.Lexer, out *ServiceBFront_Hello_Result) {
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
					out.Success = new(string)
				}
				*out.Success = string(in.String())
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
func easyjsonAfe36c91EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsMultiMultiServiceBFrontHello(out *jwriter.Writer, in ServiceBFront_Hello_Result) {
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
		out.String(string(*in.Success))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ServiceBFront_Hello_Result) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonAfe36c91EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsMultiMultiServiceBFrontHello(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ServiceBFront_Hello_Result) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonAfe36c91EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsMultiMultiServiceBFrontHello(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ServiceBFront_Hello_Result) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonAfe36c91DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsMultiMultiServiceBFrontHello(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ServiceBFront_Hello_Result) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonAfe36c91DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsMultiMultiServiceBFrontHello(l, v)
}
func easyjsonAfe36c91DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsMultiMultiServiceBFrontHello1(in *jlexer.Lexer, out *ServiceBFront_Hello_Args) {
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
func easyjsonAfe36c91EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsMultiMultiServiceBFrontHello1(out *jwriter.Writer, in ServiceBFront_Hello_Args) {
	out.RawByte('{')
	first := true
	_ = first
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ServiceBFront_Hello_Args) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonAfe36c91EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsMultiMultiServiceBFrontHello1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ServiceBFront_Hello_Args) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonAfe36c91EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsMultiMultiServiceBFrontHello1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ServiceBFront_Hello_Args) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonAfe36c91DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsMultiMultiServiceBFrontHello1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ServiceBFront_Hello_Args) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonAfe36c91DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsMultiMultiServiceBFrontHello1(l, v)
}
