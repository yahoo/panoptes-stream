package vault

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"

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
func (v *Vault) GetCredentials(path string) ([]string, error) {
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
func (v *Vault) GetCertificate(path string) (*tls.Certificate, error) {
	cert, key, err := v.GetKeyPair(path)
	if err != nil {
		return nil, err
	}

	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	return &certificate, nil
}

// GetKeyPair returns x509 key pair from Vault
func (v *Vault) GetKeyPair(path string) ([]byte, []byte, error) {
	cfg := api.DefaultConfig()
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}

	secrets, err := client.Logical().Read(path)
	if err != nil {
		return nil, nil, err
	}

	if isExist(secrets.Data, "cert") {
		return getKeyPair(secrets.Data)
	} else if isExist(secrets.Data, "pkcs12") {
		return getKeyPairPKCS12PEM(secrets.Data)
	}

	return nil, nil, errors.New("not exist")
}

// getKeyPairPKCS12PEM returns certificate from pkcs12 archive
// private key and X.509 certificate encoded as PEM
// pkcs12=pkcs12_data password=password
// password is optional
func getKeyPairPKCS12PEM(data map[string]interface{}) ([]byte, []byte, error) {
	password := ""
	b, err := base64.StdEncoding.DecodeString(data["pkcs12"].(string))
	if err != nil {
		return nil, nil, err
	}

	if isExist(data, "password") {
		password = data["password"].(string)
	}

	key, cert, err := pkcs12.Decode(b, password)
	if err != nil {
		return nil, nil, err
	}

	privateKey, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}

	keyPEM := &bytes.Buffer{}
	err = pem.Encode(keyPEM, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})
	if err != nil {
		return nil, nil, err
	}

	certPEM := &bytes.Buffer{}
	err = pem.Encode(certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	if err != nil {
		return nil, nil, err
	}

	return certPEM.Bytes(), keyPEM.Bytes(), nil
}

// getKeyPair returns private key and certificate encoded as PEM
func getKeyPair(data map[string]interface{}) ([]byte, []byte, error) {
	var key string

	cert := data["cert"].(string)

	if isExist(data, "key") {
		key = data["key"].(string)
	} else {
		key = cert
	}

	return []byte(cert), []byte(key), nil
}

func isExist(m map[string]interface{}, k string) bool {
	_, ok := m[k]
	return ok
}
