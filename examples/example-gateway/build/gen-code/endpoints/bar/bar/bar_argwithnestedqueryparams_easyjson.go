// Code generated by zanzibar
// @generated
// Checksum : sI9zn2t15j9mk+9mztlCsQ==
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

func easyjson883a4a87DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBarBarArgWithNestedQueryParams(in *jlexer.Lexer, out *Bar_ArgWithNestedQueryParams_Result) {
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
					out.Success = new(BarResponse)
				}
				easyjson883a4a87DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar(in, &*out.Success)
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
func easyjson883a4a87EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBarBarArgWithNestedQueryParams(out *jwriter.Writer, in Bar_ArgWithNestedQueryParams_Result) {
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
			easyjson883a4a87EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar(out, *in.Success)
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Bar_ArgWithNestedQueryParams_Result) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson883a4a87EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBarBarArgWithNestedQueryParams(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Bar_ArgWithNestedQueryParams_Result) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson883a4a87EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBarBarArgWithNestedQueryParams(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Bar_ArgWithNestedQueryParams_Result) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson883a4a87DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBarBarArgWithNestedQueryParams(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Bar_ArgWithNestedQueryParams_Result) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson883a4a87DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBarBarArgWithNestedQueryParams(l, v)
}
func easyjson883a4a87DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar(in *jlexer.Lexer, out *BarResponse) {
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
					out.MapIntWithRange = make(map[string]int32)
				} else {
					out.MapIntWithRange = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
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
}
func easyjson883a4a87EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar(out *jwriter.Writer, in BarResponse) {
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
		v3First := true
		for v3Name, v3Value := range in.MapIntWithRange {
			if !v3First {
				out.RawByte(',')
			}
			v3First = false
			out.String(string(v3Name))
			out.RawByte(':')
			out.Int32(int32(v3Value))
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
		v4First := true
		for v4Name, v4Value := range in.MapIntWithoutRange {
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
	out.RawByte('}')
}
func easyjson883a4a87DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBarBarArgWithNestedQueryParams1(in *jlexer.Lexer, out *Bar_ArgWithNestedQueryParams_Args) {
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
					out.Request = new(QueryParamsStruct)
				}
				easyjson883a4a87DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar1(in, &*out.Request)
			}
			RequestSet = true
		case "opt":
			if in.IsNull() {
				in.Skip()
				out.Opt = nil
			} else {
				if out.Opt == nil {
					out.Opt = new(QueryParamsOptsStruct)
				}
				easyjson883a4a87DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar2(in, &*out.Opt)
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
	if !RequestSet {
		in.AddError(fmt.Errorf("key 'request' is required"))
	}
}
func easyjson883a4a87EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBarBarArgWithNestedQueryParams1(out *jwriter.Writer, in Bar_ArgWithNestedQueryParams_Args) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"request\":")
	if in.Request == nil {
		out.RawString("null")
	} else {
		easyjson883a4a87EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar1(out, *in.Request)
	}
	if in.Opt != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"opt\":")
		if in.Opt == nil {
			out.RawString("null")
		} else {
			easyjson883a4a87EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar2(out, *in.Opt)
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Bar_ArgWithNestedQueryParams_Args) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson883a4a87EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBarBarArgWithNestedQueryParams1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Bar_ArgWithNestedQueryParams_Args) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson883a4a87EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBarBarArgWithNestedQueryParams1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Bar_ArgWithNestedQueryParams_Args) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson883a4a87DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBarBarArgWithNestedQueryParams1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Bar_ArgWithNestedQueryParams_Args) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson883a4a87DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBarBarArgWithNestedQueryParams1(l, v)
}
func easyjson883a4a87DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar2(in *jlexer.Lexer, out *QueryParamsOptsStruct) {
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
func easyjson883a4a87EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar2(out *jwriter.Writer, in QueryParamsOptsStruct) {
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
func easyjson883a4a87DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar1(in *jlexer.Lexer, out *QueryParamsStruct) {
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
func easyjson883a4a87EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeEndpointsBarBar1(out *jwriter.Writer, in QueryParamsStruct) {
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
