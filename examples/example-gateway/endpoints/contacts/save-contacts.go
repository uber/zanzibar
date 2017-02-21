package contacts

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uber-go/zap"
	"github.com/uber/zanzibar/examples/example-gateway/clients"
	contactsClientStructs "github.com/uber/zanzibar/examples/example-gateway/gen-code/github.com/uber/zanzibar/clients/contacts/contacts"
	zanzibar "github.com/uber/zanzibar/runtime"
)

// HandleSaveContactsRequest "/contacts/:userUUID/contacts"
func HandleSaveContactsRequest(
	ctx context.Context,
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
	res, err := clients.Contacts.SaveContacts(ctx, clientBody, nil)
	if err != nil {
		gateway.Logger.Error("Could not make client request",
			zap.String("error", err.Error()),
		)
		inc.SendError(500, errors.Wrap(err, "Could not make client request:"))
		return
	}

	// Handle client respnse.
	if !inc.IsOKResponse(res.StatusCode, []int{200, 202}) {
		gateway.Logger.Warn("Unknown response status code",
			zap.Int("status code", res.StatusCode),
		)
	}

	// res.Res.StatusCode
	inc.CopyJSON(res.StatusCode, res.Body)
}

func convertToClient(
	body *SaveContactsRequest,
) *contactsClientStructs.SaveContactsRequest {
	clientBody := &contactsClientStructs.SaveContactsRequest{}
	clientBody.UserUUID = contactsClientStructs.UUID(body.UserUUID)

	for _, contact := range body.Contacts {
		clientContact := &contactsClientStructs.Contact{}
		clientAttributes := &contactsClientStructs.ContactAttributes{}
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
			clientFragment := &contactsClientStructs.ContactFragment{}
			clientFragment.Text = fragment.Text
			clientFragmentType := contactsClientStructs.
				ContactFragmentType(*fragment.Type)
			clientFragment.Type = &clientFragmentType

			clientContact.Fragments = append(
				clientContact.Fragments, clientFragment,
			)
		}

		clientContact.Attributes = clientAttributes
		clientBody.Contacts = append(clientBody.Contacts, clientContact)
	}

	return clientBody
}
