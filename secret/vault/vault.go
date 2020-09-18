//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/kelseyhightower/envconfig"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

// Vault represents Hashicorp Vault
type Vault struct {
	client *api.Client
}

type vaultConfig struct {
	Address   string
	Token     string
	TLSConfig config.TLSConfig
}

// New constructs a new Vault
func New() (*Vault, error) {
	config := &vaultConfig{}
	prefix := "panoptes_vault"
	err := envconfig.Process(prefix, config)
	if err != nil {
		return nil, err
	}

	cfg := api.DefaultConfig()

	if config.Address != "" {
		cfg.Address = config.Address
	}

	if config.TLSConfig.Enabled {
		cfg.ConfigureTLS(&api.TLSConfig{
			ClientCert: config.TLSConfig.CertFile,
			ClientKey:  config.TLSConfig.KeyFile,
			CACert:     config.TLSConfig.CAFile,
			Insecure:   config.TLSConfig.InsecureSkipVerify,
		})
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	if config.Token != "" {
		client.SetToken(config.Token)
	}

	return &Vault{client: client}, nil
}

// GetSecrets returns all available data as key value for given path
// it extracts cert and private key from pkcs12 data
func (v *Vault) GetSecrets(path string) (map[string][]byte, error) {
	secrets, err := v.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("vault: %v", err)
	}

	if secrets == nil {
		return nil, fmt.Errorf("vault: path %s not exist", path)
	}

	return getSecrets(secrets.Data), nil
}

// getSecrets returns private key and certificate encoded as PEM
func getSecrets(data map[string]interface{}) map[string][]byte {
	var result = make(map[string][]byte)

	for key, value := range data {
		result[key] = []byte(value.(string))
	}

	return result
}
