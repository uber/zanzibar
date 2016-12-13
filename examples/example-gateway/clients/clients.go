package clients

import (
	"github.com/uber/zanzibar/examples/example-gateway/clients/contacts"
	"github.com/uber/zanzibar/examples/example-gateway/clients/google_now"
)

// Clients datastructure that holds all the generated clients
// This should only hold clients generate from specs
type Clients struct {
	Contacts  *contactsClient.ContactsClient
	GoogleNow *googleNow.Client
}

// Options for creating all clients
type Options struct {
	Contacts  contactsClient.Options
	GoogleNow googleNow.Options
}

// CreateClients will make all clients
func CreateClients(opts *Options) *Clients {
	return &Clients{
		Contacts:  contactsClient.Create(&opts.Contacts),
		GoogleNow: googleNow.NewClient(&opts.GoogleNow),
	}
}
