package helpers

import (
	"encoding/json"
	"github.com/hashicorp/vault/api"
)

type VaultClient struct {
	Vault      *api.Client
	DataPrefix string
	MountPoint string
}

func New(url, secretID, roleID, mountPoint string) (VaultClient, error) {
	cfg := api.Config{Address: url}
	_ = cfg.ConfigureTLS(&api.TLSConfig{Insecure: true})

	c, err := api.NewClient(&cfg)
	if err != nil {
		return VaultClient{}, err
	}

	client := VaultClient{
		Vault:      c,
		DataPrefix: mountPoint + "/data/",
		MountPoint: mountPoint,
	}

	cred := map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	resp, err := client.Vault.Logical().Write("auth/approle/login", cred)
	if err != nil || resp.Auth == nil {
		return VaultClient{}, err
	} else {
		client.Vault.SetToken(resp.Auth.ClientToken)
	}

	return client, nil
}

func (c *VaultClient) ReadSecret(path string) (interface{}, error) {
	secret, err := c.Vault.Logical().Read(c.DataPrefix + path)
	if err != nil {
		return nil, err
	}

	return secret.Data["data"], nil
}

func (c *VaultClient) ReadSecretAsBytes(path string) ([]byte, error) {
	secret, err := c.Vault.Logical().Read(c.DataPrefix + path)
	if err != nil {
		return nil, err
	}

	return json.Marshal(secret.Data["data"])
}
