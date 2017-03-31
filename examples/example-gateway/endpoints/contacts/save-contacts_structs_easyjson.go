// AUTOGENERATED FILE: easyjson marshaller/unmarshallers.

package contacts

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

func easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts(in *jlexer.Lexer, out *SaveContactsResponse) {
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
		_ = key
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
func easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts(out *jwriter.Writer, in SaveContactsResponse) {
	out.RawByte('{')
	first := true
	_ = first
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v SaveContactsResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SaveContactsResponse) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *SaveContactsResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SaveContactsResponse) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts(l, v)
}
func easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts1(in *jlexer.Lexer, out *SaveContactsRequest) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	var ContactsSet bool
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
		case "UserUUID":
			out.UserUUID = string(in.String())
		case "appType":
			out.AppType = string(in.String())
		case "contacts":
			if in.IsNull() {
				in.Skip()
				out.Contacts = nil
			} else {
				in.Delim('[')
				if !in.IsDelim(']') {
					out.Contacts = make([]*Contact, 0, 8)
				} else {
					out.Contacts = []*Contact{}
				}
				for !in.IsDelim(']') {
					var v1 *Contact
					if in.IsNull() {
						in.Skip()
						v1 = nil
					} else {
						if v1 == nil {
							v1 = new(Contact)
						}
						(*v1).UnmarshalEasyJSON(in)
					}
					out.Contacts = append(out.Contacts, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
			ContactsSet = true
		case "DeviceType":
			out.DeviceType = string(in.String())
		case "AppVersion":
			out.AppVersion = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
	if !ContactsSet {
		in.AddError(fmt.Errorf("key 'contacts' is required"))
	}
}
func easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts1(out *jwriter.Writer, in SaveContactsRequest) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"UserUUID\":")
	out.String(string(in.UserUUID))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"appType\":")
	out.String(string(in.AppType))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"contacts\":")
	if in.Contacts == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v2, v3 := range in.Contacts {
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
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"DeviceType\":")
	out.String(string(in.DeviceType))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"AppVersion\":")
	out.String(string(in.AppVersion))
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v SaveContactsRequest) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SaveContactsRequest) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *SaveContactsRequest) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SaveContactsRequest) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts1(l, v)
}
func easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts2(in *jlexer.Lexer, out *Contact) {
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
		case "fragments":
			if in.IsNull() {
				in.Skip()
				out.Fragments = nil
			} else {
				in.Delim('[')
				if !in.IsDelim(']') {
					out.Fragments = make([]*ContactFragment, 0, 8)
				} else {
					out.Fragments = []*ContactFragment{}
				}
				for !in.IsDelim(']') {
					var v4 *ContactFragment
					if in.IsNull() {
						in.Skip()
						v4 = nil
					} else {
						if v4 == nil {
							v4 = new(ContactFragment)
						}
						(*v4).UnmarshalEasyJSON(in)
					}
					out.Fragments = append(out.Fragments, v4)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "attributes":
			(out.Attributes).UnmarshalEasyJSON(in)
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
func easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts2(out *jwriter.Writer, in Contact) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"fragments\":")
	if in.Fragments == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v5, v6 := range in.Fragments {
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
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"attributes\":")
	(in.Attributes).MarshalEasyJSON(out)
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Contact) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Contact) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Contact) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Contact) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts2(l, v)
}
func easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts3(in *jlexer.Lexer, out *ContactAttributes) {
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
		case "firstName":
			if in.IsNull() {
				in.Skip()
				out.FirstName = nil
			} else {
				if out.FirstName == nil {
					out.FirstName = new(string)
				}
				*out.FirstName = string(in.String())
			}
		case "lastName":
			if in.IsNull() {
				in.Skip()
				out.LastName = nil
			} else {
				if out.LastName == nil {
					out.LastName = new(string)
				}
				*out.LastName = string(in.String())
			}
		case "nickname":
			if in.IsNull() {
				in.Skip()
				out.Nickname = nil
			} else {
				if out.Nickname == nil {
					out.Nickname = new(string)
				}
				*out.Nickname = string(in.String())
			}
		case "hasPhoto":
			if in.IsNull() {
				in.Skip()
				out.HasPhoto = nil
			} else {
				if out.HasPhoto == nil {
					out.HasPhoto = new(bool)
				}
				*out.HasPhoto = bool(in.Bool())
			}
		case "numFields":
			if in.IsNull() {
				in.Skip()
				out.NumFields = nil
			} else {
				if out.NumFields == nil {
					out.NumFields = new(int32)
				}
				*out.NumFields = int32(in.Int32())
			}
		case "timesContacted":
			if in.IsNull() {
				in.Skip()
				out.TimesContacted = nil
			} else {
				if out.TimesContacted == nil {
					out.TimesContacted = new(int32)
				}
				*out.TimesContacted = int32(in.Int32())
			}
		case "lastTimeContacted":
			if in.IsNull() {
				in.Skip()
				out.LastTimeContacted = nil
			} else {
				if out.LastTimeContacted == nil {
					out.LastTimeContacted = new(int32)
				}
				*out.LastTimeContacted = int32(in.Int32())
			}
		case "isStarred":
			if in.IsNull() {
				in.Skip()
				out.IsStarred = nil
			} else {
				if out.IsStarred == nil {
					out.IsStarred = new(bool)
				}
				*out.IsStarred = bool(in.Bool())
			}
		case "hasCustomRingtone":
			if in.IsNull() {
				in.Skip()
				out.HasCustomRingtone = nil
			} else {
				if out.HasCustomRingtone == nil {
					out.HasCustomRingtone = new(bool)
				}
				*out.HasCustomRingtone = bool(in.Bool())
			}
		case "isSendToVoiceMail":
			if in.IsNull() {
				in.Skip()
				out.IsSendToVoiceMail = nil
			} else {
				if out.IsSendToVoiceMail == nil {
					out.IsSendToVoiceMail = new(bool)
				}
				*out.IsSendToVoiceMail = bool(in.Bool())
			}
		case "hasThumbnail":
			if in.IsNull() {
				in.Skip()
				out.HasThumbnail = nil
			} else {
				if out.HasThumbnail == nil {
					out.HasThumbnail = new(bool)
				}
				*out.HasThumbnail = bool(in.Bool())
			}
		case "namePrefix":
			if in.IsNull() {
				in.Skip()
				out.NamePrefix = nil
			} else {
				if out.NamePrefix == nil {
					out.NamePrefix = new(string)
				}
				*out.NamePrefix = string(in.String())
			}
		case "nameSuffix":
			if in.IsNull() {
				in.Skip()
				out.NameSuffix = nil
			} else {
				if out.NameSuffix == nil {
					out.NameSuffix = new(string)
				}
				*out.NameSuffix = string(in.String())
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
func easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts3(out *jwriter.Writer, in ContactAttributes) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"firstName\":")
	if in.FirstName == nil {
		out.RawString("null")
	} else {
		out.String(string(*in.FirstName))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"lastName\":")
	if in.LastName == nil {
		out.RawString("null")
	} else {
		out.String(string(*in.LastName))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"nickname\":")
	if in.Nickname == nil {
		out.RawString("null")
	} else {
		out.String(string(*in.Nickname))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"hasPhoto\":")
	if in.HasPhoto == nil {
		out.RawString("null")
	} else {
		out.Bool(bool(*in.HasPhoto))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"numFields\":")
	if in.NumFields == nil {
		out.RawString("null")
	} else {
		out.Int32(int32(*in.NumFields))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"timesContacted\":")
	if in.TimesContacted == nil {
		out.RawString("null")
	} else {
		out.Int32(int32(*in.TimesContacted))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"lastTimeContacted\":")
	if in.LastTimeContacted == nil {
		out.RawString("null")
	} else {
		out.Int32(int32(*in.LastTimeContacted))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"isStarred\":")
	if in.IsStarred == nil {
		out.RawString("null")
	} else {
		out.Bool(bool(*in.IsStarred))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"hasCustomRingtone\":")
	if in.HasCustomRingtone == nil {
		out.RawString("null")
	} else {
		out.Bool(bool(*in.HasCustomRingtone))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"isSendToVoiceMail\":")
	if in.IsSendToVoiceMail == nil {
		out.RawString("null")
	} else {
		out.Bool(bool(*in.IsSendToVoiceMail))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"hasThumbnail\":")
	if in.HasThumbnail == nil {
		out.RawString("null")
	} else {
		out.Bool(bool(*in.HasThumbnail))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"namePrefix\":")
	if in.NamePrefix == nil {
		out.RawString("null")
	} else {
		out.String(string(*in.NamePrefix))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"nameSuffix\":")
	if in.NameSuffix == nil {
		out.RawString("null")
	} else {
		out.String(string(*in.NameSuffix))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ContactAttributes) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ContactAttributes) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ContactAttributes) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ContactAttributes) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts3(l, v)
}
func easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts4(in *jlexer.Lexer, out *ContactFragment) {
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
		case "type":
			if in.IsNull() {
				in.Skip()
				out.Type = nil
			} else {
				if out.Type == nil {
					out.Type = new(string)
				}
				*out.Type = string(in.String())
			}
		case "text":
			if in.IsNull() {
				in.Skip()
				out.Text = nil
			} else {
				if out.Text == nil {
					out.Text = new(string)
				}
				*out.Text = string(in.String())
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
func easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts4(out *jwriter.Writer, in ContactFragment) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"type\":")
	if in.Type == nil {
		out.RawString("null")
	} else {
		out.String(string(*in.Type))
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"text\":")
	if in.Text == nil {
		out.RawString("null")
	} else {
		out.String(string(*in.Text))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ContactFragment) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts4(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ContactFragment) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDae16842EncodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts4(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ContactFragment) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts4(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ContactFragment) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDae16842DecodeGithubComUberZanzibarExamplesExampleGatewayEndpointsContacts4(l, v)
}
