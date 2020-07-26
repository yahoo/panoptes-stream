package secret

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/secret/vault"
)

type Secret interface {
	GetSecrets(string) (map[string][]byte, error)
}

func GetSecretEngine(sType string) (Secret, error) {
	switch sType {
	case "vault":
		return vault.New()
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

func GetCredentials(sType, path string) (map[string]string, error) {
	sec, err := GetSecretEngine(sType)
	if err != nil {
		return nil, err
	}

	secrets, err := sec.GetSecrets(path)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for k, v := range secrets {
		result[k] = string(v)
	}

	return result, nil
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
	var (
		caCertPool *x509.CertPool
		tlsConfig  = &tls.Config{}
	)

	sType, path, ok := ParseRemoteSecretInfo(cfg.CertFile)
	if !ok {
		return nil, false, nil
	}

	sec, err := GetSecretEngine(sType)
	if err != nil {
		return nil, ok, err
	}

	secrets, err := sec.GetSecrets(path)
	if err != nil {
		return nil, ok, err
	}

	if !isExist(secrets, "cert") && !isExist(secrets, "ca") {
		return nil, ok, errors.New("secrets are not available")
	}

	if isExist(secrets, "cert") {
		if !isExist(secrets, "key") {
			secrets["key"] = secrets["cert"]
		}

		cert, err := tls.X509KeyPair(secrets["cert"], secrets["key"])
		if err != nil {
			return nil, ok, err
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if isExist(secrets, "ca") {
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(secrets["ca"])

		tlsConfig.RootCAs = caCertPool
	}

	tlsConfig.InsecureSkipVerify = cfg.InsecureSkipVerify
	tlsConfig.Renegotiation = tls.RenegotiateNever

	return tlsConfig, ok, nil

}

func getTLSConfigLocal(cfg *config.TLSConfig) (*tls.Config, error) {
	var (
		caCertPool *x509.CertPool
		tlsConfig  = &tls.Config{}
	)

	if cfg.CertFile != "" {
		if cfg.KeyFile == "" {
			cfg.KeyFile = cfg.CertFile
		}

		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, err
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if cfg.CAFile != "" {
		caCert, err := ioutil.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, err
		}

		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig.RootCAs = caCertPool
	}

	tlsConfig.InsecureSkipVerify = cfg.InsecureSkipVerify
	tlsConfig.Renegotiation = tls.RenegotiateNever

	return tlsConfig, nil
}

func isExist(m map[string][]byte, k string) bool {
	_, ok := m[k]
	return ok
}
