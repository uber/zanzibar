// Code generated by thriftrw v1.9.0. DO NOT EDIT.
// @generated

package contacts

import (
	"errors"
	"fmt"
	"go.uber.org/thriftrw/wire"
	"strings"
)

type Contact struct {
	Fragments  []*ContactFragment `json:"fragments,omitempty"`
	Attributes *ContactAttributes `json:"attributes,omitempty"`
}

type _List_ContactFragment_ValueList []*ContactFragment

func (v _List_ContactFragment_ValueList) ForEach(f func(wire.Value) error) error {
	for i, x := range v {
		if x == nil {
			return fmt.Errorf("invalid [%v]: value is nil", i)
		}
		w, err := x.ToWire()
		if err != nil {
			return err
		}
		err = f(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v _List_ContactFragment_ValueList) Size() int {
	return len(v)
}

func (_List_ContactFragment_ValueList) ValueType() wire.Type {
	return wire.TStruct
}

func (_List_ContactFragment_ValueList) Close() {}

// ToWire translates a Contact struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//   x, err := v.ToWire()
//   if err != nil {
//     return err
//   }
//
//   if err := binaryProtocol.Encode(x, writer); err != nil {
//     return err
//   }
func (v *Contact) ToWire() (wire.Value, error) {
	var (
		fields [2]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Fragments != nil {
		w, err = wire.NewValueList(_List_ContactFragment_ValueList(v.Fragments)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}
	if v.Attributes != nil {
		w, err = v.Attributes.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 2, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _ContactFragment_Read(w wire.Value) (*ContactFragment, error) {
	var v ContactFragment
	err := v.FromWire(w)
	return &v, err
}

func _List_ContactFragment_Read(l wire.ValueList) ([]*ContactFragment, error) {
	if l.ValueType() != wire.TStruct {
		return nil, nil
	}

	o := make([]*ContactFragment, 0, l.Size())
	err := l.ForEach(func(x wire.Value) error {
		i, err := _ContactFragment_Read(x)
		if err != nil {
			return err
		}
		o = append(o, i)
		return nil
	})
	l.Close()
	return o, err
}

func _ContactAttributes_Read(w wire.Value) (*ContactAttributes, error) {
	var v ContactAttributes
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a Contact struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a Contact struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v Contact
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *Contact) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TList {
				v.Fragments, err = _List_ContactFragment_Read(field.Value.GetList())
				if err != nil {
					return err
				}

			}
		case 2:
			if field.Value.Type() == wire.TStruct {
				v.Attributes, err = _ContactAttributes_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a Contact
// struct.
func (v *Contact) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [2]string
	i := 0
	if v.Fragments != nil {
		fields[i] = fmt.Sprintf("Fragments: %v", v.Fragments)
		i++
	}
	if v.Attributes != nil {
		fields[i] = fmt.Sprintf("Attributes: %v", v.Attributes)
		i++
	}

	return fmt.Sprintf("Contact{%v}", strings.Join(fields[:i], ", "))
}

func _List_ContactFragment_Equals(lhs, rhs []*ContactFragment) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i, lv := range lhs {
		rv := rhs[i]
		if !lv.Equals(rv) {
			return false
		}
	}

	return true
}

// Equals returns true if all the fields of this Contact match the
// provided Contact.
//
// This function performs a deep comparison.
func (v *Contact) Equals(rhs *Contact) bool {
	if !((v.Fragments == nil && rhs.Fragments == nil) || (v.Fragments != nil && rhs.Fragments != nil && _List_ContactFragment_Equals(v.Fragments, rhs.Fragments))) {
		return false
	}
	if !((v.Attributes == nil && rhs.Attributes == nil) || (v.Attributes != nil && rhs.Attributes != nil && v.Attributes.Equals(rhs.Attributes))) {
		return false
	}

	return true
}

type ContactAttributes struct {
	FirstName         *string `json:"firstName,omitempty"`
	LastName          *string `json:"lastName,omitempty"`
	Nickname          *string `json:"nickname,omitempty"`
	HasPhoto          *bool   `json:"hasPhoto,omitempty"`
	NumFields         *int32  `json:"numFields,omitempty"`
	TimesContacted    *int32  `json:"timesContacted,omitempty"`
	LastTimeContacted *int32  `json:"lastTimeContacted,omitempty"`
	IsStarred         *bool   `json:"isStarred,omitempty"`
	HasCustomRingtone *bool   `json:"hasCustomRingtone,omitempty"`
	IsSendToVoicemail *bool   `json:"isSendToVoicemail,omitempty"`
	HasThumbnail      *bool   `json:"hasThumbnail,omitempty"`
	NamePrefix        *string `json:"namePrefix,omitempty"`
	NameSuffix        *string `json:"nameSuffix,omitempty"`
}

// ToWire translates a ContactAttributes struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//   x, err := v.ToWire()
//   if err != nil {
//     return err
//   }
//
//   if err := binaryProtocol.Encode(x, writer); err != nil {
//     return err
//   }
func (v *ContactAttributes) ToWire() (wire.Value, error) {
	var (
		fields [13]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.FirstName != nil {
		w, err = wire.NewValueString(*(v.FirstName)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}
	if v.LastName != nil {
		w, err = wire.NewValueString(*(v.LastName)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 2, Value: w}
		i++
	}
	if v.Nickname != nil {
		w, err = wire.NewValueString(*(v.Nickname)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 3, Value: w}
		i++
	}
	if v.HasPhoto != nil {
		w, err = wire.NewValueBool(*(v.HasPhoto)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 4, Value: w}
		i++
	}
	if v.NumFields != nil {
		w, err = wire.NewValueI32(*(v.NumFields)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 5, Value: w}
		i++
	}
	if v.TimesContacted != nil {
		w, err = wire.NewValueI32(*(v.TimesContacted)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 6, Value: w}
		i++
	}
	if v.LastTimeContacted != nil {
		w, err = wire.NewValueI32(*(v.LastTimeContacted)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 7, Value: w}
		i++
	}
	if v.IsStarred != nil {
		w, err = wire.NewValueBool(*(v.IsStarred)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 8, Value: w}
		i++
	}
	if v.HasCustomRingtone != nil {
		w, err = wire.NewValueBool(*(v.HasCustomRingtone)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 9, Value: w}
		i++
	}
	if v.IsSendToVoicemail != nil {
		w, err = wire.NewValueBool(*(v.IsSendToVoicemail)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 10, Value: w}
		i++
	}
	if v.HasThumbnail != nil {
		w, err = wire.NewValueBool(*(v.HasThumbnail)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 11, Value: w}
		i++
	}
	if v.NamePrefix != nil {
		w, err = wire.NewValueString(*(v.NamePrefix)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 12, Value: w}
		i++
	}
	if v.NameSuffix != nil {
		w, err = wire.NewValueString(*(v.NameSuffix)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 13, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a ContactAttributes struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a ContactAttributes struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v ContactAttributes
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *ContactAttributes) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.FirstName = &x
				if err != nil {
					return err
				}

			}
		case 2:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.LastName = &x
				if err != nil {
					return err
				}

			}
		case 3:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.Nickname = &x
				if err != nil {
					return err
				}

			}
		case 4:
			if field.Value.Type() == wire.TBool {
				var x bool
				x, err = field.Value.GetBool(), error(nil)
				v.HasPhoto = &x
				if err != nil {
					return err
				}

			}
		case 5:
			if field.Value.Type() == wire.TI32 {
				var x int32
				x, err = field.Value.GetI32(), error(nil)
				v.NumFields = &x
				if err != nil {
					return err
				}

			}
		case 6:
			if field.Value.Type() == wire.TI32 {
				var x int32
				x, err = field.Value.GetI32(), error(nil)
				v.TimesContacted = &x
				if err != nil {
					return err
				}

			}
		case 7:
			if field.Value.Type() == wire.TI32 {
				var x int32
				x, err = field.Value.GetI32(), error(nil)
				v.LastTimeContacted = &x
				if err != nil {
					return err
				}

			}
		case 8:
			if field.Value.Type() == wire.TBool {
				var x bool
				x, err = field.Value.GetBool(), error(nil)
				v.IsStarred = &x
				if err != nil {
					return err
				}

			}
		case 9:
			if field.Value.Type() == wire.TBool {
				var x bool
				x, err = field.Value.GetBool(), error(nil)
				v.HasCustomRingtone = &x
				if err != nil {
					return err
				}

			}
		case 10:
			if field.Value.Type() == wire.TBool {
				var x bool
				x, err = field.Value.GetBool(), error(nil)
				v.IsSendToVoicemail = &x
				if err != nil {
					return err
				}

			}
		case 11:
			if field.Value.Type() == wire.TBool {
				var x bool
				x, err = field.Value.GetBool(), error(nil)
				v.HasThumbnail = &x
				if err != nil {
					return err
				}

			}
		case 12:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.NamePrefix = &x
				if err != nil {
					return err
				}

			}
		case 13:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.NameSuffix = &x
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a ContactAttributes
// struct.
func (v *ContactAttributes) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [13]string
	i := 0
	if v.FirstName != nil {
		fields[i] = fmt.Sprintf("FirstName: %v", *(v.FirstName))
		i++
	}
	if v.LastName != nil {
		fields[i] = fmt.Sprintf("LastName: %v", *(v.LastName))
		i++
	}
	if v.Nickname != nil {
		fields[i] = fmt.Sprintf("Nickname: %v", *(v.Nickname))
		i++
	}
	if v.HasPhoto != nil {
		fields[i] = fmt.Sprintf("HasPhoto: %v", *(v.HasPhoto))
		i++
	}
	if v.NumFields != nil {
		fields[i] = fmt.Sprintf("NumFields: %v", *(v.NumFields))
		i++
	}
	if v.TimesContacted != nil {
		fields[i] = fmt.Sprintf("TimesContacted: %v", *(v.TimesContacted))
		i++
	}
	if v.LastTimeContacted != nil {
		fields[i] = fmt.Sprintf("LastTimeContacted: %v", *(v.LastTimeContacted))
		i++
	}
	if v.IsStarred != nil {
		fields[i] = fmt.Sprintf("IsStarred: %v", *(v.IsStarred))
		i++
	}
	if v.HasCustomRingtone != nil {
		fields[i] = fmt.Sprintf("HasCustomRingtone: %v", *(v.HasCustomRingtone))
		i++
	}
	if v.IsSendToVoicemail != nil {
		fields[i] = fmt.Sprintf("IsSendToVoicemail: %v", *(v.IsSendToVoicemail))
		i++
	}
	if v.HasThumbnail != nil {
		fields[i] = fmt.Sprintf("HasThumbnail: %v", *(v.HasThumbnail))
		i++
	}
	if v.NamePrefix != nil {
		fields[i] = fmt.Sprintf("NamePrefix: %v", *(v.NamePrefix))
		i++
	}
	if v.NameSuffix != nil {
		fields[i] = fmt.Sprintf("NameSuffix: %v", *(v.NameSuffix))
		i++
	}

	return fmt.Sprintf("ContactAttributes{%v}", strings.Join(fields[:i], ", "))
}

func _String_EqualsPtr(lhs, rhs *string) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

func _Bool_EqualsPtr(lhs, rhs *bool) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

func _I32_EqualsPtr(lhs, rhs *int32) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

// Equals returns true if all the fields of this ContactAttributes match the
// provided ContactAttributes.
//
// This function performs a deep comparison.
func (v *ContactAttributes) Equals(rhs *ContactAttributes) bool {
	if !_String_EqualsPtr(v.FirstName, rhs.FirstName) {
		return false
	}
	if !_String_EqualsPtr(v.LastName, rhs.LastName) {
		return false
	}
	if !_String_EqualsPtr(v.Nickname, rhs.Nickname) {
		return false
	}
	if !_Bool_EqualsPtr(v.HasPhoto, rhs.HasPhoto) {
		return false
	}
	if !_I32_EqualsPtr(v.NumFields, rhs.NumFields) {
		return false
	}
	if !_I32_EqualsPtr(v.TimesContacted, rhs.TimesContacted) {
		return false
	}
	if !_I32_EqualsPtr(v.LastTimeContacted, rhs.LastTimeContacted) {
		return false
	}
	if !_Bool_EqualsPtr(v.IsStarred, rhs.IsStarred) {
		return false
	}
	if !_Bool_EqualsPtr(v.HasCustomRingtone, rhs.HasCustomRingtone) {
		return false
	}
	if !_Bool_EqualsPtr(v.IsSendToVoicemail, rhs.IsSendToVoicemail) {
		return false
	}
	if !_Bool_EqualsPtr(v.HasThumbnail, rhs.HasThumbnail) {
		return false
	}
	if !_String_EqualsPtr(v.NamePrefix, rhs.NamePrefix) {
		return false
	}
	if !_String_EqualsPtr(v.NameSuffix, rhs.NameSuffix) {
		return false
	}

	return true
}

// GetFirstName returns the value of FirstName if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetFirstName() (o string) {
	if v.FirstName != nil {
		return *v.FirstName
	}

	return
}

// GetLastName returns the value of LastName if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetLastName() (o string) {
	if v.LastName != nil {
		return *v.LastName
	}

	return
}

// GetNickname returns the value of Nickname if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetNickname() (o string) {
	if v.Nickname != nil {
		return *v.Nickname
	}

	return
}

// GetHasPhoto returns the value of HasPhoto if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetHasPhoto() (o bool) {
	if v.HasPhoto != nil {
		return *v.HasPhoto
	}

	return
}

// GetNumFields returns the value of NumFields if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetNumFields() (o int32) {
	if v.NumFields != nil {
		return *v.NumFields
	}

	return
}

// GetTimesContacted returns the value of TimesContacted if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetTimesContacted() (o int32) {
	if v.TimesContacted != nil {
		return *v.TimesContacted
	}

	return
}

// GetLastTimeContacted returns the value of LastTimeContacted if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetLastTimeContacted() (o int32) {
	if v.LastTimeContacted != nil {
		return *v.LastTimeContacted
	}

	return
}

// GetIsStarred returns the value of IsStarred if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetIsStarred() (o bool) {
	if v.IsStarred != nil {
		return *v.IsStarred
	}

	return
}

// GetHasCustomRingtone returns the value of HasCustomRingtone if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetHasCustomRingtone() (o bool) {
	if v.HasCustomRingtone != nil {
		return *v.HasCustomRingtone
	}

	return
}

// GetIsSendToVoicemail returns the value of IsSendToVoicemail if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetIsSendToVoicemail() (o bool) {
	if v.IsSendToVoicemail != nil {
		return *v.IsSendToVoicemail
	}

	return
}

// GetHasThumbnail returns the value of HasThumbnail if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetHasThumbnail() (o bool) {
	if v.HasThumbnail != nil {
		return *v.HasThumbnail
	}

	return
}

// GetNamePrefix returns the value of NamePrefix if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetNamePrefix() (o string) {
	if v.NamePrefix != nil {
		return *v.NamePrefix
	}

	return
}

// GetNameSuffix returns the value of NameSuffix if it is set or its
// zero value if it is unset.
func (v *ContactAttributes) GetNameSuffix() (o string) {
	if v.NameSuffix != nil {
		return *v.NameSuffix
	}

	return
}

type ContactFragment struct {
	Type *ContactFragmentType `json:"type,omitempty"`
	Text *string              `json:"text,omitempty"`
}

// ToWire translates a ContactFragment struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//   x, err := v.ToWire()
//   if err != nil {
//     return err
//   }
//
//   if err := binaryProtocol.Encode(x, writer); err != nil {
//     return err
//   }
func (v *ContactFragment) ToWire() (wire.Value, error) {
	var (
		fields [2]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Type != nil {
		w, err = v.Type.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}
	if v.Text != nil {
		w, err = wire.NewValueString(*(v.Text)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 2, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _ContactFragmentType_Read(w wire.Value) (ContactFragmentType, error) {
	var x ContactFragmentType
	err := x.FromWire(w)
	return x, err
}

// FromWire deserializes a ContactFragment struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a ContactFragment struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v ContactFragment
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *ContactFragment) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				var x ContactFragmentType
				x, err = _ContactFragmentType_Read(field.Value)
				v.Type = &x
				if err != nil {
					return err
				}

			}
		case 2:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.Text = &x
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a ContactFragment
// struct.
func (v *ContactFragment) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [2]string
	i := 0
	if v.Type != nil {
		fields[i] = fmt.Sprintf("Type: %v", *(v.Type))
		i++
	}
	if v.Text != nil {
		fields[i] = fmt.Sprintf("Text: %v", *(v.Text))
		i++
	}

	return fmt.Sprintf("ContactFragment{%v}", strings.Join(fields[:i], ", "))
}

func _ContactFragmentType_EqualsPtr(lhs, rhs *ContactFragmentType) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

// Equals returns true if all the fields of this ContactFragment match the
// provided ContactFragment.
//
// This function performs a deep comparison.
func (v *ContactFragment) Equals(rhs *ContactFragment) bool {
	if !_ContactFragmentType_EqualsPtr(v.Type, rhs.Type) {
		return false
	}
	if !_String_EqualsPtr(v.Text, rhs.Text) {
		return false
	}

	return true
}

// GetType returns the value of Type if it is set or its
// zero value if it is unset.
func (v *ContactFragment) GetType() (o ContactFragmentType) {
	if v.Type != nil {
		return *v.Type
	}

	return
}

// GetText returns the value of Text if it is set or its
// zero value if it is unset.
func (v *ContactFragment) GetText() (o string) {
	if v.Text != nil {
		return *v.Text
	}

	return
}

type ContactFragmentType string

// ToWire translates ContactFragmentType into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
func (v ContactFragmentType) ToWire() (wire.Value, error) {
	x := (string)(v)
	return wire.NewValueString(x), error(nil)
}

// String returns a readable string representation of ContactFragmentType.
func (v ContactFragmentType) String() string {
	x := (string)(v)
	return fmt.Sprint(x)
}

// FromWire deserializes ContactFragmentType from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
func (v *ContactFragmentType) FromWire(w wire.Value) error {
	x, err := w.GetString(), error(nil)
	*v = (ContactFragmentType)(x)
	return err
}

// Equals returns true if this ContactFragmentType is equal to the provided
// ContactFragmentType.
func (lhs ContactFragmentType) Equals(rhs ContactFragmentType) bool {
	return (lhs == rhs)
}

type SaveContactsRequest struct {
	UserUUID string     `json:"userUUID,required"`
	Contacts []*Contact `json:"contacts,required"`
}

type _List_Contact_ValueList []*Contact

func (v _List_Contact_ValueList) ForEach(f func(wire.Value) error) error {
	for i, x := range v {
		if x == nil {
			return fmt.Errorf("invalid [%v]: value is nil", i)
		}
		w, err := x.ToWire()
		if err != nil {
			return err
		}
		err = f(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v _List_Contact_ValueList) Size() int {
	return len(v)
}

func (_List_Contact_ValueList) ValueType() wire.Type {
	return wire.TStruct
}

func (_List_Contact_ValueList) Close() {}

// ToWire translates a SaveContactsRequest struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//   x, err := v.ToWire()
//   if err != nil {
//     return err
//   }
//
//   if err := binaryProtocol.Encode(x, writer); err != nil {
//     return err
//   }
func (v *SaveContactsRequest) ToWire() (wire.Value, error) {
	var (
		fields [2]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueString(v.UserUUID), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++
	if v.Contacts == nil {
		return w, errors.New("field Contacts of SaveContactsRequest is required")
	}
	w, err = wire.NewValueList(_List_Contact_ValueList(v.Contacts)), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 2, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _Contact_Read(w wire.Value) (*Contact, error) {
	var v Contact
	err := v.FromWire(w)
	return &v, err
}

func _List_Contact_Read(l wire.ValueList) ([]*Contact, error) {
	if l.ValueType() != wire.TStruct {
		return nil, nil
	}

	o := make([]*Contact, 0, l.Size())
	err := l.ForEach(func(x wire.Value) error {
		i, err := _Contact_Read(x)
		if err != nil {
			return err
		}
		o = append(o, i)
		return nil
	})
	l.Close()
	return o, err
}

// FromWire deserializes a SaveContactsRequest struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a SaveContactsRequest struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v SaveContactsRequest
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *SaveContactsRequest) FromWire(w wire.Value) error {
	var err error

	userUUIDIsSet := false
	contactsIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.UserUUID, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				userUUIDIsSet = true
			}
		case 2:
			if field.Value.Type() == wire.TList {
				v.Contacts, err = _List_Contact_Read(field.Value.GetList())
				if err != nil {
					return err
				}
				contactsIsSet = true
			}
		}
	}

	if !userUUIDIsSet {
		return errors.New("field UserUUID of SaveContactsRequest is required")
	}

	if !contactsIsSet {
		return errors.New("field Contacts of SaveContactsRequest is required")
	}

	return nil
}

// String returns a readable string representation of a SaveContactsRequest
// struct.
func (v *SaveContactsRequest) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [2]string
	i := 0
	fields[i] = fmt.Sprintf("UserUUID: %v", v.UserUUID)
	i++
	fields[i] = fmt.Sprintf("Contacts: %v", v.Contacts)
	i++

	return fmt.Sprintf("SaveContactsRequest{%v}", strings.Join(fields[:i], ", "))
}

func _List_Contact_Equals(lhs, rhs []*Contact) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i, lv := range lhs {
		rv := rhs[i]
		if !lv.Equals(rv) {
			return false
		}
	}

	return true
}

// Equals returns true if all the fields of this SaveContactsRequest match the
// provided SaveContactsRequest.
//
// This function performs a deep comparison.
func (v *SaveContactsRequest) Equals(rhs *SaveContactsRequest) bool {
	if !(v.UserUUID == rhs.UserUUID) {
		return false
	}
	if !_List_Contact_Equals(v.Contacts, rhs.Contacts) {
		return false
	}

	return true
}

type SaveContactsResponse struct {
}

// ToWire translates a SaveContactsResponse struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//   x, err := v.ToWire()
//   if err != nil {
//     return err
//   }
//
//   if err := binaryProtocol.Encode(x, writer); err != nil {
//     return err
//   }
func (v *SaveContactsResponse) ToWire() (wire.Value, error) {
	var (
		fields [0]wire.Field
		i      int = 0
	)

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a SaveContactsResponse struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a SaveContactsResponse struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v SaveContactsResponse
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *SaveContactsResponse) FromWire(w wire.Value) error {

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		}
	}

	return nil
}

// String returns a readable string representation of a SaveContactsResponse
// struct.
func (v *SaveContactsResponse) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [0]string
	i := 0

	return fmt.Sprintf("SaveContactsResponse{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this SaveContactsResponse match the
// provided SaveContactsResponse.
//
// This function performs a deep comparison.
func (v *SaveContactsResponse) Equals(rhs *SaveContactsResponse) bool {

	return true
}

type UUID string

// ToWire translates UUID into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
func (v UUID) ToWire() (wire.Value, error) {
	x := (string)(v)
	return wire.NewValueString(x), error(nil)
}

// String returns a readable string representation of UUID.
func (v UUID) String() string {
	x := (string)(v)
	return fmt.Sprint(x)
}

// FromWire deserializes UUID from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
func (v *UUID) FromWire(w wire.Value) error {
	x, err := w.GetString(), error(nil)
	*v = (UUID)(x)
	return err
}

// Equals returns true if this UUID is equal to the provided
// UUID.
func (lhs UUID) Equals(rhs UUID) bool {
	return (lhs == rhs)
}
