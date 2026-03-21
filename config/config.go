// Package config
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	NewRelicLicenseKey string
	BucketName         string
	TimeParseLayout    string
	NewRelicURL        string
	NewRelicAccountID  string
}

func (c *Config) Load(path string) (*Config, error) {
	configOptions := Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &configOptions)
	if err != nil {
		return nil, err
	}

	return &configOptions, nil
}

func (c *Config) validate() error {
	if err := requireField("NewRelicLicenseKey", c.NewRelicLicenseKey); err != nil {
		return err
	}
	if err := requireField("BucketName", c.BucketName); err != nil {
		return err
	}
	if err := requireField("TimeParseLayout", c.TimeParseLayout); err != nil {
		c.TimeParseLayout = "2006-01-02 15:04:05"
	}

	return nil
}

func requireField(name, inputField string) error {
	if inputField == "" {
		return fmt.Errorf("required field [%s] not found in config", name)
	}
	return nil
}
