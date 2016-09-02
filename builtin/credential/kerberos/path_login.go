package kerberos

import (
	"fmt"
	"strings"

	"github.com/hashicorp/vault/helper/policyutil"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathLogin(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "NULL", //TODO: Need to figure out Pattern
		Fields: map[string]*framework.FieldSchema{
			"ticket": &framework.FieldSchema{
				Type:        framework.TypeString, //TODO: Figure out if this is right
				Description: "Kerberos Ticket used to Authorize and Authenticate",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathLogin,
		},

		HelpSynopsis:    pathLoginSyn,
		HelpDescription: pathLoginDesc,
	}
}

func (b *backend) pathLogin(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

}

func (b *backend) pathLoginRenew(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

}

const pathLoginSyn = `
Placeholder
`

const pathLoginDesc = `
Placeholder
`
