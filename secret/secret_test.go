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

	pem.Encode(buf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	privateKeyPem := buf.String()

	caFile, err := ioutil.TempFile("", "caFile")
	assert.NoError(t, err)
	defer os.Remove(caFile.Name())
	caFile.WriteString(certPem)

	keyFile, err := ioutil.TempFile("", "keyFile")
	assert.NoError(t, err)
	defer os.Remove(keyFile.Name())
	keyFile.WriteString(privateKeyPem)

	cfg := &config.TLSConfig{
		CertFile: caFile.Name(),
		KeyFile:  keyFile.Name(),
	}

	tlsCfg, err := getTLSConfigLocal(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, tlsCfg)
}
