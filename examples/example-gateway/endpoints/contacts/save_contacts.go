package contacts

import (
	"context"

	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/contacts/module"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/contacts/workflow"
	contactsClientStructs "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/contacts/contacts"
	endpointContacts "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/contacts/contacts"
	"github.com/uber/zanzibar/runtime"

	"go.uber.org/zap"
)

// NewContactsSaveContactsWorkflow ...
func NewContactsSaveContactsWorkflow(
	deps *module.Dependencies,
) workflow.ContactsSaveContactsWorkflow {
	return &SaveContactsEndpoint{
		Clients: deps.Client,
		Logger:  deps.Default.Logger,
	}
}

// SaveContactsEndpoint ...
type SaveContactsEndpoint struct {
	Clients *module.ClientDependencies
	Logger  *zap.Logger
	Request *zanzibar.ServerHTTPRequest
}

// Handle "/contacts/:userUUID/contacts"
func (w SaveContactsEndpoint) Handle(
	ctx context.Context,
	headers zanzibar.Header,
	r *endpointContacts.SaveContactsRequest,
) (*endpointContacts.SaveContactsResponse, zanzibar.Header, error) {
	// TODO AuthenticatedRequest()
	// TODO MatchedIdRequest({paramName: 'userUUID'})

	clientBody := convertToClient(r)
	cres, _, err := w.Clients.Contacts.SaveContacts(ctx, nil, clientBody)
	if err != nil {
		w.Logger.Error("Could not make client request", zap.Error(err))
		return nil, nil, err
	}

	// TODO: verify IsOKResponse() on client response status code

	response := convertToResponse(cres)
	return response, nil, nil
}

func convertToResponse(
	body *contactsClientStructs.SaveContactsResponse,
) *endpointContacts.SaveContactsResponse {
	return &endpointContacts.SaveContactsResponse{}
}

func convertToClient(
	body *endpointContacts.SaveContactsRequest,
) *contactsClientStructs.SaveContactsRequest {
	clientBody := &contactsClientStructs.SaveContactsRequest{}
	clientBody.UserUUID = body.UserUUID

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
		clientAttributes.IsSendToVoicemail = attributes.IsSendToVoicemail
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
