package clients

import (
	"github.com/uber/zanzibar/examples/example-gateway/clients/bar"
	"github.com/uber/zanzibar/examples/example-gateway/clients/contacts"
	"github.com/uber/zanzibar/examples/example-gateway/clients/googlenow"
	"github.com/uber/zanzibar/runtime"
)

// Clients datastructure that holds all the generated clients
// This should only hold clients generate from specs
type Clients struct {
	Contacts  *contactsClient.ContactsClient
	GoogleNow *googlenowClient.GoogleNowClient
	Bar       *barClient.BarClient
}

// CreateClients will make all clients
func CreateClients(config *zanzibar.StaticConfig) *Clients {
	return &Clients{
		Contacts:  contactsClient.NewClient(config),
		GoogleNow: googlenowClient.NewClient(config),
		Bar:       barClient.NewClient(config),
	}
}
