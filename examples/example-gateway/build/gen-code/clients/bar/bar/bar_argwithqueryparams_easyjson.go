// Code generated by zanzibar
// @generated
// Checksum : yl4Dj83zjKrzKKFnZE+HUA==
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

func easyjsonCec46174DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarArgWithQueryParams(in *jlexer.Lexer, out *Bar_ArgWithQueryParams_Result) {
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
				easyjsonCec46174DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar(in, &*out.Success)
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
func easyjsonCec46174EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarArgWithQueryParams(out *jwriter.Writer, in Bar_ArgWithQueryParams_Result) {
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
			easyjsonCec46174EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar(out, *in.Success)
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Bar_ArgWithQueryParams_Result) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonCec46174EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarArgWithQueryParams(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Bar_ArgWithQueryParams_Result) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonCec46174EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarArgWithQueryParams(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Bar_ArgWithQueryParams_Result) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonCec46174DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarArgWithQueryParams(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Bar_ArgWithQueryParams_Result) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonCec46174DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarArgWithQueryParams(l, v)
}
func easyjsonCec46174DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar(in *jlexer.Lexer, out *BarResponse) {
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
		case "resp":
			if in.IsNull() {
				in.Skip()
				out.Resp = nil
			} else {
				if out.Resp == nil {
					out.Resp = new(BarRequestRecur)
				}
				easyjsonCec46174DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar1(in, &*out.Resp)
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
func easyjsonCec46174EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar(out *jwriter.Writer, in BarResponse) {
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
	if in.Resp != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"resp\":")
		if in.Resp == nil {
			out.RawString("null")
		} else {
			easyjsonCec46174EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar1(out, *in.Resp)
		}
	}
	out.RawByte('}')
}
func easyjsonCec46174DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar1(in *jlexer.Lexer, out *BarRequestRecur) {
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
		case "recur":
			if in.IsNull() {
				in.Skip()
				out.Recur = nil
			} else {
				if out.Recur == nil {
					out.Recur = new(BarRequestRecur)
				}
				easyjsonCec46174DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar1(in, &*out.Recur)
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
func easyjsonCec46174EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar1(out *jwriter.Writer, in BarRequestRecur) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"name\":")
	out.String(string(in.Name))
	if in.Recur != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"recur\":")
		if in.Recur == nil {
			out.RawString("null")
		} else {
			easyjsonCec46174EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBar1(out, *in.Recur)
		}
	}
	out.RawByte('}')
}
func easyjsonCec46174DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarArgWithQueryParams1(in *jlexer.Lexer, out *Bar_ArgWithQueryParams_Args) {
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
func easyjsonCec46174EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarArgWithQueryParams1(out *jwriter.Writer, in Bar_ArgWithQueryParams_Args) {
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
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Bar_ArgWithQueryParams_Args) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonCec46174EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarArgWithQueryParams1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Bar_ArgWithQueryParams_Args) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonCec46174EncodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarArgWithQueryParams1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Bar_ArgWithQueryParams_Args) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonCec46174DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarArgWithQueryParams1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Bar_ArgWithQueryParams_Args) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonCec46174DecodeGithubComUberZanzibarExamplesExampleGatewayBuildGenCodeClientsBarBarBarArgWithQueryParams1(l, v)
}
