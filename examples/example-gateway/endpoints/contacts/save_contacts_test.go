package contacts_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	mockcontactsworkflow "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/contacts/mock-workflow"
	endpointContacts "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/contacts/contacts"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
)

func TestSaveContactsCall(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	ms.MockClients().Contacts.ExpectSaveContacts().Success()

	endpointRequest := &endpointContacts.Contacts_SaveContacts_Args{
		SaveContactsRequest: &endpointContacts.SaveContactsRequest{
			Contacts: []*endpointContacts.Contact{},
		},
	}
	rawBody, _ := endpointRequest.MarshalJSON()

	res, err := ms.MakeHTTPRequest(
		"POST", "/contacts/foo/contacts", nil, bytes.NewReader(rawBody),
	)

	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "202 Accepted", res.Status)
}

func TestSaveContactsCallWorkflow(t *testing.T) {
	mh, mc := mockcontactsworkflow.NewContactsSaveContactsWorkflowMock(t)

	mc.Contacts.ExpectSaveContacts().Success()

	endpointRequest := &endpointContacts.Contacts_SaveContacts_Args{
		SaveContactsRequest: &endpointContacts.SaveContactsRequest{
			UserUUID: "foo",
			Contacts: []*endpointContacts.Contact{},
		},
	}

	_, res, resHeaders, err := mh.Handle(context.Background(), nil, endpointRequest)

	if !assert.NoError(t, err, "got error") {
		return
	}
	assert.Nil(t, resHeaders)
	assert.Equal(t, &endpointContacts.SaveContactsResponse{}, res)
}
