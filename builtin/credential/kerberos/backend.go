package kerberos

import (
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func Factory(conf *logical.BackendConfig) (logical.Backend, error) {
	return Backend().Setup(conf)
}

func Backend() *backend {

	var b backend

	return &b
}

type backend struct {
	*framework.Backend
}

const backendHelp = `
Text
`
