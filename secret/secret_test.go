//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package secret

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"

	kv "github.com/hashicorp/vault-plugin-secrets-kv"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
	"github.com/stretchr/testify/assert"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

func TestGetTLSConfigLocal(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"panoptes"},
		},

		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 1),

		DNSNames: []string{"not-exist.com"},

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derCertBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: derCertBytes})
	certPem := buf.String()

	buf.Reset()

	pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: derCertBytes})
	caPem := buf.String()

	buf.Reset()

	pem.Encode(buf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	privateKeyPem := buf.String()

	certFile, err := ioutil.TempFile("", "certFile")
	assert.NoError(t, err)
	defer os.Remove(certFile.Name())
	certFile.WriteString(certPem)

	caFile, err := ioutil.TempFile("", "caFile")
	assert.NoError(t, err)
	defer os.Remove(caFile.Name())
	caFile.WriteString(caPem)

	keyFile, err := ioutil.TempFile("", "keyFile")
	assert.NoError(t, err)
	defer os.Remove(keyFile.Name())
	keyFile.WriteString(privateKeyPem)

	cfg := &config.TLSConfig{
		CertFile: certFile.Name(),
		KeyFile:  keyFile.Name(),
		CAFile:   caFile.Name(),
	}

	tlsCfg, err := GetTLSConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, tlsCfg)

	cfg = &config.TLSConfig{
		CertFile: "notexist",
		KeyFile:  "notexist",
		CAFile:   "notexist",
	}

	tlsCfg, err = GetTLSConfig(cfg)
	assert.Error(t, err)
	assert.Nil(t, tlsCfg)
}

func TestParseRemoteSecretInfo(t *testing.T) {
	typ, path, ok := ParseRemoteSecretInfo("__vault::path")
	assert.Equal(t, "vault", typ)
	assert.Equal(t, "path", path)
	assert.True(t, ok)

	typ, path, ok = ParseRemoteSecretInfo("notremote")
	assert.Equal(t, "", typ)
	assert.Equal(t, "", path)
	assert.False(t, ok)
}

func TestGetTLSConfigRemote(t *testing.T) {
	os.Setenv("PANOPTES_VAULT_TLSCONFIG_ENABLED", "true")
	os.Setenv("PANOPTES_VAULT_TLSCONFIG_INSECURESKIPVERIFY", "true")

	cluster := createVaultTestCluster(t)
	defer cluster.Cleanup()

	client := cluster.Cores[0].Client

	os.Setenv("PANOPTES_VAULT_TOKEN", client.Token())

	////// put private key and cert
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.NoError(t, err)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"panoptes"},
		},

		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 1),

		DNSNames: []string{"not-exist.com"},

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derCertBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	assert.NoError(t, err)
	buf := &bytes.Buffer{}
	pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: derCertBytes})
	certPem := buf.String()
	buf.Reset()
	pem.Encode(buf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	keyPem := buf.String()
	path := "secrets/v1/tls"
	data := map[string]interface{}{"key": keyPem, "cert": certPem}

	client.Logical().Write(path, data)
	//////

	cfg := &config.TLSConfig{
		CertFile: "__vault::secrets/v1/tls",
	}

	tlscfg, err := GetTLSConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, tlscfg)
}

func TestGetTLSServerConfig(t *testing.T) {
	cfg := &config.TLSConfig{}
	_, err := GetTLSServerConfig(cfg)
	assert.Error(t, err)
}

func TestGetCredentials(t *testing.T) {
	os.Setenv("PANOPTES_VAULT_TLSCONFIG_ENABLED", "true")
	os.Setenv("PANOPTES_VAULT_TLSCONFIG_INSECURESKIPVERIFY", "true")

	cluster := createVaultTestCluster(t)
	defer cluster.Cleanup()

	client := cluster.Cores[0].Client

	os.Setenv("PANOPTES_VAULT_TOKEN", client.Token())

	path := "secrets/v1/creds"
	data := map[string]interface{}{"token": "topsecret"}
	client.Logical().Write(path, data)

	secrets, err := GetCredentials("vault", "secrets/v1/creds")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"token": "topsecret"}, secrets)

	_, err = GetCredentials("notexist", "secrets/v1/creds")
	assert.Error(t, err)

	_, err = GetCredentials("vault", "notexist")
	assert.Error(t, err)
}

func createVaultTestCluster(t *testing.T) *vault.TestCluster {
	t.Helper()

	coreConfig := &vault.CoreConfig{
		LogicalBackends: map[string]logical.Factory{
			"kv": kv.Factory,
		},
	}

	cluster := vault.NewTestCluster(t, coreConfig, &vault.TestClusterOptions{
		HandlerFunc:       http.Handler,
		NumCores:          1,
		Logger:            nil,
		BaseListenAddress: "127.0.0.1:8200",
		RequireClientAuth: false,
	})

	cluster.Start()

	if err := cluster.Cores[0].Client.Sys().Mount("/secrets/v1", &api.MountInput{
		Type: "kv",
		Options: map[string]string{
			"path": "/secrets/v1",
		},
	}); err != nil {
		t.Fatal(err)
	}

	return cluster
}
