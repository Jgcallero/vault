package kerberos

import (
	"fmt"
	//"strings"

	//"github.com/hashicorp/vault/helper/policyutil"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathKeytab(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keytab", //TODO: Check to see if this needs changing
		Fields: map[string]*framework.FieldSchema{
			"keytab": &framework.FieldSchema{
				Type:        framework.TypeString, //This seems wrong
				Description: "Keytab to authenticate users",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.DeleteOperation: b.pathKeytabDelete,
			logical.ReadOperation:   b.pathKeytabRead,
			logical.UpdateOperation: b.pathKeytabWrite,
		},

		HelpSynopsis:    pathKeytabHelpSyn,
		HelpDescription: pathKeytabHelpDesc,
	}
}

func (b *backend) pathKeytabDelete(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	fmt.Println("Keytab Delete")
	return logical.ErrorResponse("Not Supported Yet"), nil

}

func (b *backend) pathKeytabRead(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	fmt.Println("Keytab Read")
	return logical.ErrorResponse("Not Supported Yet"), nil

}

func (b *backend) pathKeytabWrite(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	fmt.Println("Keytab Write")
	return logical.ErrorResponse("Not Supported Yet"), nil

}

const pathKeytabHelpSyn = `
Placeholder
`

const pathKeytabHelpDesc = `
Placeholder
`
