package contactsClient

//go:generate easyjson -all $GOFILE

// ContactFragment ...
type ContactFragment struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ContactAttributes ...
type ContactAttributes struct {
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Nickname          string `json:"nickname"`
	HasPhoto          bool   `json:"has_photo"`
	NumFields         int32  `json:"num_fields"`
	TimesContacted    int32  `json:"times_contacted"`
	LastTimeContacted int32  `json:"last_time_contacted"`
	IsStarred         bool   `json:"is_starred"`
	HasCustomRingtone bool   `json:"has_custom_ringtone"`
	IsSendToVoicemail bool   `json:"is_send_to_voicemail"`
	HasThumbnail      bool   `json:"has_thumbnail"`
	NamePrefix        string `json:"name_prefix"`
	NameSuffix        string `json:"name_suffix"`
}

// Contact ...
type Contact struct {
	Fragments  []*ContactFragment `json:"fragments"`
	Attributes ContactAttributes  `json:"attributes"`
}

// SaveContactsRequest ...
type SaveContactsRequest struct {
	UserUUID   string     `json:"-"`
	AppType    string     `json:"app_type"`
	Contacts   []*Contact `json:"contacts"`
	DeviceType string     `json:"device_type"`
	AppVersion string     `json:"app_version"`
}
