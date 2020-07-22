package vault

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"reflect"

	"github.com/hashicorp/vault/api"
	"software.sslmate.com/src/go-pkcs12"
)

// Vault represents Hashicorp Vault
type Vault struct {
}

// New constructs a new Vault
func New() *Vault {
	return &Vault{}
}

// GetCredentials returns username and password from Vault
// format: username=password at specified path
func (v *Vault) GetCredentials(ctx context.Context, path string) ([]string, error) {
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
		return []string{k, v.(string)}, nil
	}

	return nil, errors.New("credentials not found")
}

// GetCertificate returns TLS certificate from Vault
func (v *Vault) GetCertificate(ctx context.Context, path string) (*tls.Certificate, error) {
	cfg := api.DefaultConfig()
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	secrets, err := client.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	keys := reflect.ValueOf(secrets.Data).MapKeys()
	if len(keys) < 1 {

	}

	switch {
	case isExist(secrets.Data, "pkcs12"):
		return pkcs12pem(secrets.Data)

	}

	return nil, errors.New("not exist")
}

// pkcs12pem returns certificate from pkcs12 archive
// private key and X.509 certificate encoded as PEM
// pkcs12=pkcs12_data password=password
// password is optional
func pkcs12pem(data map[string]interface{}) (*tls.Certificate, error) {
	password := ""
	b, err := base64.StdEncoding.DecodeString(data["pkcs12"].(string))
	if err != nil {
		return nil, err
	}

	if isExist(data, "password") {
		password = data["password"].(string)
	}

	key, cert, err := pkcs12.Decode(b, password)
	if err != nil {
		return nil, err
	}

	privateKey, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}

	keyPEM := &bytes.Buffer{}
	err = pem.Encode(keyPEM, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})
	if err != nil {
		return nil, err
	}

	certPEM := &bytes.Buffer{}
	err = pem.Encode(certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	if err != nil {
		return nil, err
	}

	certificate, err := tls.X509KeyPair(certPEM.Bytes(), keyPEM.Bytes())
	if err != nil {
		return nil, err
	}

	return &certificate, nil
}

func isExist(m map[string]interface{}, k string) bool {
	_, ok := m[k]
	return ok
}
