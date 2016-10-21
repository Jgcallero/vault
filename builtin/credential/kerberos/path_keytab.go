package kerberos

import (
	"fmt"
	"strings"

	//"github.com/hashicorp/vault/helper/policyutil"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathKeytabWrite(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keytab/write", //TODO: Check to see if this needs changing
		Fields: map[string]*framework.FieldSchema{
			"principal": &framework.FieldSchema{
				Type:        framework.TypeString, //This seems wrong
				Description: "Principal for the user",
			},

			"realm": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Realm for the user",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.DeleteOperation: b.pathKeytabDelete,
			logical.UpdateOperation: b.pathKeytabWrite,
		},

		HelpSynopsis:    pathKeytabHelpSyn,
		HelpDescription: pathKeytabHelpDesc,
	}
}

func pathKeytabRead(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keytab/read",
		Fields: map[string]*framework.FieldSchema{
			"principal": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Principal for the user",
			},

			"realm": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "realm for the user",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathKeytabRead,
		},

		HelpSynopsis:    pathKeytabHelpSyn,
		HelpDescription: pathKeytabHelpDesc,
	}
}

func (b *backend) pathKeytabDelete(req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	fmt.Println("Keytab Delete")
	return logical.ErrorResponse("Not Supported Yet"), nil

}

func (b *backend) pathKeytabRead(req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	fmt.Println("Keytab Read")
	Princ := strings.ToLower(d.Get("principal").(string))
	Realm := strings.ToUpper(d.Get("realm").(string))
	fmt.Println("keytab/" + Princ + Realm)
	entry, err := req.Storage.Get("keytab/" + Princ + Realm)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return logical.ErrorResponse("prinicipal/realm doesn't exist"), nil
	}
	Ticket := &TicketEntry{}
	if err := entry.DecodeJSON(&Ticket); err != nil {
		return nil, err
	}

	fmt.Println(Ticket.Principal, ' ', Ticket.Realm)
	return logical.ErrorResponse("We Did It"), nil
}

func (b *backend) pathKeytabWrite(req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	fmt.Println("Keytab Write")
	Principal := strings.ToLower(d.Get("principal").(string))
	Ticket := &TicketEntry{}
	Ticket.Principal = Principal
	Realm := strings.ToUpper(d.Get("realm").(string))
	Ticket.Realm = Realm
	fmt.Println("keytab/" + Principal + Realm)
	entry, err := logical.StorageEntryJSON("keytab/"+Principal+Realm, Ticket)
	if err != nil {
		return nil, err
	}

	return nil, req.Storage.Put(entry)

}

type TicketEntry struct {
	//Principal is the Principal or name used
	//For the user
	Principal string

	//Realm is the realm orAuthentication Administrative
	//Domain
	Realm string

	//TODO: Add more sections

}

const pathKeytabHelpSyn = `
Placeholder
`

const pathKeytabHelpDesc = `
Placeholder
`
