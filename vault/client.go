package vault

import (
	"encoding/json"
	"github.com/hashicorp/vault/api"
)

type Client struct {
	Vault      *api.Client
	DataPrefix string
	MountPoint string
}

func New(url, secretID, roleID, mountPoint string) (Client, error) {
	cfg := api.Config{Address: url}
	_ = cfg.ConfigureTLS(&api.TLSConfig{Insecure: true})

	c, err := api.NewClient(&cfg)
	if err != nil {
		return Client{}, err
	}

	client := Client{
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
		return Client{}, err
	} else {
		client.Vault.SetToken(resp.Auth.ClientToken)
	}

	return client, nil
}

func (c *Client) ReadSecret(path string) (interface{}, error) {
	secret, err := c.Vault.Logical().Read(c.DataPrefix + path)
	if err != nil {
		return nil, err
	}

	return secret.Data["data"], nil
}

func (c *Client) ReadSecretAsBytes(path string) ([]byte, error) {
	secret, err := c.Vault.Logical().Read(c.DataPrefix + path)
	if err != nil {
		return nil, err
	}

	return json.Marshal(secret.Data["data"])
}
