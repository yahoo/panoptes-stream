//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package vault

import (
	"testing"

	kv "github.com/hashicorp/vault-plugin-secrets-kv"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
)

func createVaultTestCluster(t *testing.T) *vault.TestCluster {
	t.Helper()

	coreConfig := &vault.CoreConfig{
		LogicalBackends: map[string]logical.Factory{
			"kv": kv.Factory,
		},
	}
	cluster := vault.NewTestCluster(t, coreConfig, &vault.TestClusterOptions{
		HandlerFunc: http.Handler,
		NumCores:    1,
		Logger:      nil,
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

func TestGetSecrets(t *testing.T) {
	cluster := createVaultTestCluster(t)
	defer cluster.Cleanup()

	client := cluster.Cores[0].Client

	path := "secrets/v1/influxdb"
	data := map[string]interface{}{"token": "topsecret"}

	client.Logical().Write(path, data)

	v := &Vault{client}
	r, err := v.GetSecrets(path)
	if err != nil {
		t.Fatal(err)
	}

	secret, ok := r["token"]
	if !ok {
		t.Fatal("expect to have a secret but got", string(secret))
	}

	if string(secret) != "topsecret" {
		t.Error("expect to get secret: topsecret but got", string(secret))
	}
}
