package contacts

import (
	"context"

	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/uber-go/zap"
	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	contactsClientStructs "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/clients/contacts/contacts"
	zanzibar "github.com/uber/zanzibar/runtime"
)

// HandleSaveContactsRequest "/contacts/:userUUID/contacts"
func HandleSaveContactsRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
	clients *clients.Clients,
) {
	var body SaveContactsRequest
	ok := req.ReadAndUnmarshalBody(&body)
	if !ok {
		return
	}

	// TODO AuthenticatedRequest()
	// TODO MatchedIdRequest({paramName: 'userUUID'})

	body.UserUUID = req.Params[0].Value
	body.AppType = req.Header.Get("x-uber-client-name")
	body.DeviceType = req.Header.Get("x-uber-device")
	body.AppVersion = req.Header.Get("x-uber-client-version")

	clientBody := convertToClient(&body)
	cres, err := clients.Contacts.SaveContacts(ctx, clientBody, nil)
	if err != nil {
		req.Logger.Error("Could not make client request",
			zap.String("error", err.Error()),
		)
		res.SendError(500, errors.Wrap(err, "Could not make client request:"))
		res.Flush()
		return
	}

	defer func() {
		if cerr := cres.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	// Handle client respnse.
	if !res.IsOKResponse(cres.StatusCode, []int{200, 202}) {
		req.Logger.Warn("Unknown response status code",
			zap.Int("status code", cres.StatusCode),
		)
	}

	bytes, err := ioutil.ReadAll(cres.Body)
	if err != nil {
		res.SendError(500, errors.Wrap(err, "could not read client response body:"))
		res.Flush()
		return
	}
	var clientRespBody contactsClientStructs.SaveContactsResponse
	if err := clientRespBody.UnmarshalJSON(bytes); err != nil {
		res.SendError(500, errors.Wrap(err, "could not unmarshal client response body:"))
		res.Flush()
		return
	}
	response := convertToResponse(&clientRespBody)
	res.WriteJSON(cres.StatusCode, response)
	res.Flush()
}

func convertToResponse(
	body *contactsClientStructs.SaveContactsResponse,
) *SaveContactsResponse {
	return &SaveContactsResponse{}
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
