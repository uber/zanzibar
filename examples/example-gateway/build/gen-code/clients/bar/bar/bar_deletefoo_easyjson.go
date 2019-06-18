// Code generated by zanzibar
// @generated
// Checksum : Bvv1ub8mv75ECFPJdjomXA==
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

func easyjson68ad7c03DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarDeleteFoo(in *jlexer.Lexer, out *Bar_DeleteFoo_Result) {
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
func easyjson68ad7c03EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarDeleteFoo(out *jwriter.Writer, in Bar_DeleteFoo_Result) {
	out.RawByte('{')
	first := true
	_ = first
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Bar_DeleteFoo_Result) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson68ad7c03EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarDeleteFoo(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Bar_DeleteFoo_Result) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson68ad7c03EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarDeleteFoo(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Bar_DeleteFoo_Result) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson68ad7c03DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarDeleteFoo(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Bar_DeleteFoo_Result) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson68ad7c03DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarDeleteFoo(l, v)
}
func easyjson68ad7c03DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarDeleteFoo1(in *jlexer.Lexer, out *Bar_DeleteFoo_Args) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var RequestSet bool
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
		case "request":
			if in.IsNull() {
				in.Skip()
				out.Request = nil
			} else {
				if out.Request == nil {
					out.Request = new(BarRequest)
				}
				easyjson68ad7c03DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar(in, out.Request)
			}
			RequestSet = true
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !RequestSet {
		in.AddError(fmt.Errorf("key 'request' is required"))
	}
}
func easyjson68ad7c03EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarDeleteFoo1(out *jwriter.Writer, in Bar_DeleteFoo_Args) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"request\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Request == nil {
			out.RawString("null")
		} else {
			easyjson68ad7c03EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar(out, *in.Request)
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Bar_DeleteFoo_Args) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson68ad7c03EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarDeleteFoo1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Bar_DeleteFoo_Args) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson68ad7c03EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarDeleteFoo1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Bar_DeleteFoo_Args) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson68ad7c03DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarDeleteFoo1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Bar_DeleteFoo_Args) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson68ad7c03DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarDeleteFoo1(l, v)
}
func easyjson68ad7c03DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar(in *jlexer.Lexer, out *BarRequest) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var StringFieldSet bool
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
			out.StringField = string(in.String())
			StringFieldSet = true
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
	if !StringFieldSet {
		in.AddError(fmt.Errorf("key 'stringField' is required"))
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
func easyjson68ad7c03EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar(out *jwriter.Writer, in BarRequest) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"stringField\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.StringField))
	}
	{
		const prefix string = ",\"boolField\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Bool(bool(in.BoolField))
	}
	{
		const prefix string = ",\"binaryField\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Base64Bytes(in.BinaryField)
	}
	{
		const prefix string = ",\"timestamp\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Raw((in.Timestamp).MarshalJSON())
	}
	{
		const prefix string = ",\"enumField\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Raw((in.EnumField).MarshalJSON())
	}
	{
		const prefix string = ",\"longField\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Raw((in.LongField).MarshalJSON())
	}
	out.RawByte('}')
}
