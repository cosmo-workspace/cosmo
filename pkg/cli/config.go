package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Endpoint          string `json:"endpoint"`
	Token             string `json:"token"`
	User              string `json:"user"`
	UseServiceAccount bool   `json:"useServiceAccount,omitempty"`
	CACert            string `json:"cacert,omitempty"`

	cfg string
}

func (c *Config) GetCACert() []byte {
	if c.CACert == "" {
		return []byte{}
	}
	s, err := base64.StdEncoding.DecodeString(c.CACert)
	if err != nil {
		return []byte{}
	}
	return s
}

func (c *Config) SetCACert(ca []byte) {
	c.CACert = base64.StdEncoding.EncodeToString(ca)
}

func (c *Config) SetPath(path string) {
	c.cfg = path
}

func NewOrLoadConfigFile(path string) (*Config, error) {
	var cfg Config
	cfg.SetPath(path)

	f, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &cfg, nil
		}
		return &cfg, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	err = json.Unmarshal(f, &cfg)
	if err != nil {
		return &cfg, fmt.Errorf("invalid config file: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	b, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marchal JSON: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(c.cfg), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(c.cfg), err)
	}

	err = os.WriteFile(c.cfg, b, 0600)
	if err != nil {
		return fmt.Errorf("failed to create config file %s: %w", c.cfg, err)
	}
	return nil
}
