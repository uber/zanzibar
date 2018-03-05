package contacts_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	endpointContacts "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/contacts/contacts"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
)

func TestSaveContactsCall(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	ms.MockClientNodes().Contacts.ExpectSaveContacts().Success()

	endpointReqeust := &endpointContacts.SaveContactsRequest{
		Contacts: []*endpointContacts.Contact{},
	}
	rawBody, _ := endpointReqeust.MarshalJSON()

	res, err := ms.MakeHTTPRequest(
		"POST", "/contacts/foo/contacts", nil, bytes.NewReader(rawBody),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "202 Accepted", res.Status)
}
