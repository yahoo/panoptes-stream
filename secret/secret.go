package secret

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"regexp"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/secret/vault"
)

type Secret interface {
	GetCredentials(string) ([]string, error)
	GetCertificate(string) (*tls.Certificate, error)
	GetKeyPair(string) ([]byte, []byte, error)
}

func GetSecretEngine(sType string) (Secret, error) {
	switch sType {
	case "vault":
		return vault.New(), nil
	}

	return nil, fmt.Errorf("%s secret engine doesn't support", sType)
}

func GetTLSConfig(cfg *config.TLSConfig) (*tls.Config, error) {
	tlsConfig, ok, err := getTLSConfigRemote(cfg)
	if ok {
		return tlsConfig, err
	}

	return getTLSConfigLocal(cfg)
}

func ParseRemoteSecretInfo(key string) (string, string, bool) {
	re := regexp.MustCompile(`__([a-zA-Z0-9]*)::(.*)`)
	match := re.FindStringSubmatch(key)
	if len(match) < 1 {
		return "", "", false
	}

	return match[1], match[2], true
}

func getTLSConfigRemote(cfg *config.TLSConfig) (*tls.Config, bool, error) {
	sType, path, ok := ParseRemoteSecretInfo(cfg.CertFile)
	if ok {
		sec, err := GetSecretEngine(sType)
		if err != nil {
			return nil, ok, err
		}

		cert, err := sec.GetCertificate(path)
		if err != nil {
			return nil, ok, err
		}

		return &tls.Config{
			Certificates:       []tls.Certificate{*cert},
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		}, ok, nil
	}

	return nil, false, nil
}

func getTLSConfigLocal(cfg *config.TLSConfig) (*tls.Config, error) {
	var caCertPool *x509.CertPool

	// combined cert and private key
	if len(cfg.KeyFile) < 1 {
		cfg.KeyFile = cfg.CertFile
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, err
	}

	if cfg.CAFile != "" {
		caCert, err := ioutil.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, err
		}

		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}, nil
}
