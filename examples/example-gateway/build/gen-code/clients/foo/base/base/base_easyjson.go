// Code generated by zanzibar
// @generated
// Checksum : HEc6VVM7VSZFugsjh182PA==
// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package base

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

func easyjson25720c23DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsFooBaseBase(in *jlexer.Lexer, out *Message) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var BodySet bool
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "body":
			out.Body = string(in.String())
			BodySet = true
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !BodySet {
		in.AddError(fmt.Errorf("key 'body' is required"))
	}
}
func easyjson25720c23EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsFooBaseBase(out *jwriter.Writer, in Message) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"body\":"
		out.RawString(prefix[1:])
		out.String(string(in.Body))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Message) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson25720c23EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsFooBaseBase(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Message) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson25720c23EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsFooBaseBase(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Message) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson25720c23DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsFooBaseBase(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Message) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson25720c23DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsFooBaseBase(l, v)
}
