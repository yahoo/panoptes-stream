package secret

import (
	"context"
	"crypto/tls"
	"fmt"

	"git.vzbuilders.com/marshadrad/panoptes/secret/vault"
)

type Secret interface {
	GetCredentials(context.Context, string) ([]string, error)
	GetCertificate(context.Context, string) (*tls.Certificate, error)
}

func GetSecretEngine(sType string) (Secret, error) {
	switch sType {
	case "vault":
		return vault.New(), nil
	}

	return nil, fmt.Errorf("%s secret engine doesn't support", sType)
}
