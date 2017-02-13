package clients

import (
	"github.com/uber/zanzibar/examples/example-gateway/clients/bar"
	"github.com/uber/zanzibar/examples/example-gateway/clients/contacts"
	"github.com/uber/zanzibar/examples/example-gateway/clients/google_now"
	"github.com/uber/zanzibar/runtime"
)

// Clients datastructure that holds all the generated clients
// This should only hold clients generate from specs
type Clients struct {
	Contacts  *contactsClient.ContactsClient
	GoogleNow *googleNow.Client
	Bar       *barClient.BarClient
}

// Options for creating all clients
type Options struct {
	Contacts  contactsClient.Options
	GoogleNow googleNow.Options
	Bar       zanzibar.HTTPClientOptions
}

// CreateClients will make all clients
func CreateClients(opts *Options) *Clients {
	return &Clients{
		Contacts:  contactsClient.Create(&opts.Contacts),
		GoogleNow: googleNow.NewClient(&opts.GoogleNow),
		Bar:       barClient.NewClient(&opts.Bar),
	}
}
