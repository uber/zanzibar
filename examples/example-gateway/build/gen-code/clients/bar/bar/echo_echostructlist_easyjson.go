// Code generated by zanzibar
// @generated
// Checksum : XKTiDntP7+tKO5rJrX/JaQ==
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

func easyjson5ea6cdc8DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarEchoEchoStructList(in *jlexer.Lexer, out *Echo_EchoStructList_Result) {
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
				in.Delim('[')
				if out.Success == nil {
					if !in.IsDelim(']') {
						out.Success = make([]*BarResponse, 0, 8)
					} else {
						out.Success = []*BarResponse{}
					}
				} else {
					out.Success = (out.Success)[:0]
				}
				for !in.IsDelim(']') {
					var v1 *BarResponse
					if in.IsNull() {
						in.Skip()
						v1 = nil
					} else {
						if v1 == nil {
							v1 = new(BarResponse)
						}
						(*v1).UnmarshalEasyJSON(in)
					}
					out.Success = append(out.Success, v1)
					in.WantComma()
				}
				in.Delim(']')
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
func easyjson5ea6cdc8EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarEchoEchoStructList(out *jwriter.Writer, in Echo_EchoStructList_Result) {
	out.RawByte('{')
	first := true
	_ = first
	if len(in.Success) != 0 {
		const prefix string = ",\"success\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		{
			out.RawByte('[')
			for v2, v3 := range in.Success {
				if v2 > 0 {
					out.RawByte(',')
				}
				if v3 == nil {
					out.RawString("null")
				} else {
					(*v3).MarshalEasyJSON(out)
				}
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Echo_EchoStructList_Result) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson5ea6cdc8EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarEchoEchoStructList(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Echo_EchoStructList_Result) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson5ea6cdc8EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarEchoEchoStructList(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Echo_EchoStructList_Result) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson5ea6cdc8DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarEchoEchoStructList(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Echo_EchoStructList_Result) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson5ea6cdc8DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarEchoEchoStructList(l, v)
}
func easyjson5ea6cdc8DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarEchoEchoStructList1(in *jlexer.Lexer, out *Echo_EchoStructList_Args) {
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
				in.Delim('[')
				if out.Arg == nil {
					if !in.IsDelim(']') {
						out.Arg = make([]*BarResponse, 0, 8)
					} else {
						out.Arg = []*BarResponse{}
					}
				} else {
					out.Arg = (out.Arg)[:0]
				}
				for !in.IsDelim(']') {
					var v4 *BarResponse
					if in.IsNull() {
						in.Skip()
						v4 = nil
					} else {
						if v4 == nil {
							v4 = new(BarResponse)
						}
						(*v4).UnmarshalEasyJSON(in)
					}
					out.Arg = append(out.Arg, v4)
					in.WantComma()
				}
				in.Delim(']')
			}
			ArgSet = true
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
func easyjson5ea6cdc8EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarEchoEchoStructList1(out *jwriter.Writer, in Echo_EchoStructList_Args) {
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
		if in.Arg == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v5, v6 := range in.Arg {
				if v5 > 0 {
					out.RawByte(',')
				}
				if v6 == nil {
					out.RawString("null")
				} else {
					(*v6).MarshalEasyJSON(out)
				}
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Echo_EchoStructList_Args) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson5ea6cdc8EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarEchoEchoStructList1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Echo_EchoStructList_Args) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson5ea6cdc8EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarEchoEchoStructList1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Echo_EchoStructList_Args) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson5ea6cdc8DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarEchoEchoStructList1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Echo_EchoStructList_Args) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson5ea6cdc8DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarEchoEchoStructList1(l, v)
}
