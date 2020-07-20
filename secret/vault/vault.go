package vault

import (
	"context"
	"crypto/tls"
	"errors"

	"github.com/hashicorp/vault/api"

	"git.vzbuilders.com/marshadrad/panoptes/secret"
)

type Vault struct {
}

func New() *Vault {
	return &Vault{}
}

func (v *Vault) GetCredentials(ctx context.Context, path string) (*secret.Credentials, error) {
	cfg := api.DefaultConfig()
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	secrets, err := client.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	for k, v := range secrets.Data {
		return &secret.Credentials{
			Username: k,
			Password: v.(string),
		}, nil
	}

	return nil, errors.New("credentials not found")
}

func (v *Vault) GetCertificate(ctx context.Context, path string) (*tls.Certificate, error) {
	return nil, nil
}
