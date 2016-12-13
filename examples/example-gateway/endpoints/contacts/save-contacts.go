package contacts

//go:generate easyjson -all $GOFILE

import (
	errors "github.com/pkg/errors"
	"github.com/uber-go/zap"
	"github.com/uber/zanzibar/examples/example-gateway/clients"
	contactsClient "github.com/uber/zanzibar/examples/example-gateway/clients/contacts"
	zanzibar "github.com/uber/zanzibar/runtime"
)

// HandleSaveContactsRequest "/contacts/:userUUID/contacts"
func HandleSaveContactsRequest(
	inc *zanzibar.IncomingMessage,
	gateway *zanzibar.Gateway,
	clients *clients.Clients,
) {
	rawBody, ok := inc.ReadAll()
	if !ok {
		return
	}

	var body SaveContactsRequest
	ok = inc.UnmarshalBody(&body, rawBody)
	if !ok {
		return
	}

	// TODO AuthenticatedRequest()
	// TODO MatchedIdRequest({paramName: 'userUUID'})

	body.UserUUID = inc.Params[0].Value
	body.AppType = inc.Header.Get("x-uber-client-name")
	body.DeviceType = inc.Header.Get("x-uber-device")
	body.AppVersion = inc.Header.Get("x-uber-client-version")

	clientBody := convertToClient(&body)
	res, err := clients.Contacts.SaveContacts(clientBody)
	if err != nil {
		gateway.Logger.Error("Could not make client request",
			zap.String("error", err.Error()),
		)
		inc.SendError(500, errors.Wrap(err, "Could not make client request:"))
		return
	}

	// res.Res.StatusCode
	inc.CopyJSON(res.Res.StatusCode, res.Res.Body)
}

func convertToClient(
	body *SaveContactsRequest,
) *contactsClient.SaveContactsRequest {
	clientBody := &contactsClient.SaveContactsRequest{}
	clientBody.AppType = body.AppType
	clientBody.AppVersion = body.AppVersion
	clientBody.DeviceType = body.DeviceType
	clientBody.UserUUID = body.UserUUID

	for _, contact := range body.Contacts {
		clientContact := &contactsClient.Contact{}
		clientAttributes := clientContact.Attributes
		attributes := contact.Attributes

		clientAttributes.FirstName = attributes.FirstName
		clientAttributes.LastName = attributes.LastName
		clientAttributes.Nickname = attributes.Nickname
		clientAttributes.HasPhoto = attributes.HasPhoto
		clientAttributes.NumFields = attributes.NumFields
		clientAttributes.TimesContacted = attributes.TimesContacted
		clientAttributes.LastTimeContacted = attributes.LastTimeContacted
		clientAttributes.IsStarred = attributes.IsStarred
		clientAttributes.HasCustomRingtone = attributes.HasCustomRingtone
		clientAttributes.IsSendToVoicemail = attributes.IsSendToVoiceMail
		clientAttributes.HasThumbnail = attributes.HasThumbnail
		clientAttributes.NamePrefix = attributes.NamePrefix
		clientAttributes.NameSuffix = attributes.NameSuffix

		for _, fragment := range contact.Fragments {
			clientFragment := &contactsClient.ContactFragment{}
			clientFragment.Text = fragment.Text
			clientFragment.Type = fragment.Type

			clientContact.Fragments = append(
				clientContact.Fragments, clientFragment,
			)
		}

		clientBody.Contacts = append(clientBody.Contacts, clientContact)
	}

	return clientBody
}
