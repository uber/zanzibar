package contacts

import (
	"context"
	"time"

	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/contacts/module"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/contacts/workflow"
	contactsClientStructs "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/contacts/contacts"
	endpointContacts "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/contacts/contacts"
	zanzibar "github.com/uber/zanzibar/runtime"

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
	r *endpointContacts.Contacts_SaveContacts_Args,
) (context.Context, *endpointContacts.SaveContactsResponse, zanzibar.Header, error) {
	// TODO AuthenticatedRequest()
	// TODO MatchedIdRequest({paramName: 'userUUID'})

	clientBody := convertToClient(r)
	_, cres, _, err := w.Clients.Contacts.SaveContacts(ctx, nil, clientBody, &zanzibar.TimeoutAndRetryOptions{
		OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
		RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
		MaxAttempts:                  2,
		BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
	})
	if err != nil {
		w.Logger.Error("Could not make client request", zap.Error(err))
		return ctx, nil, nil, err
	}

	// TODO: verify IsOKResponse() on client response status code

	response := convertToResponse(cres)
	return ctx, response, nil, nil
}

func convertToResponse(
	body *contactsClientStructs.SaveContactsResponse,
) *endpointContacts.SaveContactsResponse {
	return &endpointContacts.SaveContactsResponse{}
}

func convertToClient(
	body *endpointContacts.Contacts_SaveContacts_Args,
) *contactsClientStructs.Contacts_SaveContacts_Args {
	ret := &contactsClientStructs.Contacts_SaveContacts_Args{}
	clientBody := &contactsClientStructs.SaveContactsRequest{}

	for _, contact := range body.SaveContactsRequest.Contacts {
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

	ret.SaveContactsRequest = clientBody
	ret.SaveContactsRequest.UserUUID = body.SaveContactsRequest.UserUUID
	return ret
}
