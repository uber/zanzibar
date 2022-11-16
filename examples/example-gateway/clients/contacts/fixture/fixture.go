package fixture

import (
	mc "github.com/uber/zanzibar/examples/example-gateway/build/clients/contacts/mock-client"
	gen "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/contacts/contacts"
)

var saveContactsFixtures = &mc.SaveContactsScenarios{
	Success: &mc.SaveContactsFixture{
		Arg0Any: true,
		Arg1Any: true,
		Arg2: &gen.Contacts_SaveContacts_Args{
			SaveContactsRequest: &gen.SaveContactsRequest{
				UserUUID: "foo",
			},
		},
		Arg3Any: true,

		Ret1: &gen.SaveContactsResponse{},
	},
}

// Fixture ...
var Fixture = &mc.ClientFixture{
	SaveContacts: saveContactsFixtures,
}
