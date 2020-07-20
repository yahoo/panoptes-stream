package secret

import (
	"context"
	"crypto/tls"
)

type Credentials struct {
	Username string
	Password string
}

type Secret interface {
	GetCredentials(context.Context, string) (*Credentials, error)
	GetCertificate(context.Context, string) (*tls.Certificate, error)
}
