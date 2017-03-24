package contacts

// ContactFragment ...
type ContactFragment struct {
	Type *string `json:"type"`
	Text *string `json:"text"`
}

// ContactAttributes ...
type ContactAttributes struct {
	FirstName         *string `json:"firstName"`
	LastName          *string `json:"lastName"`
	Nickname          *string `json:"nickname"`
	HasPhoto          *bool   `json:"hasPhoto"`
	NumFields         *int32  `json:"numFields"`
	TimesContacted    *int32  `json:"timesContacted"`
	LastTimeContacted *int32  `json:"lastTimeContacted"`
	IsStarred         *bool   `json:"isStarred"`
	HasCustomRingtone *bool   `json:"hasCustomRingtone"`
	IsSendToVoiceMail *bool   `json:"isSendToVoiceMail"`
	HasThumbnail      *bool   `json:"hasThumbnail"`
	NamePrefix        *string `json:"namePrefix"`
	NameSuffix        *string `json:"nameSuffix"`
}

// Contact ...
type Contact struct {
	Fragments  []*ContactFragment `json:"fragments"`
	Attributes ContactAttributes  `json:"attributes"`
}

// SaveContactsRequest ...
type SaveContactsRequest struct {
	UserUUID   string
	AppType    string     `json:"appType"`
	Contacts   []*Contact `json:"contacts,required"`
	DeviceType string
	AppVersion string
}

// SaveContactsResponse ...
type SaveContactsResponse struct {
}
