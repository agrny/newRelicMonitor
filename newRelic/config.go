package newrelic

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	NewRelicLicenseKey string `json:"newreliclicensekey"`
	BucketName         string `json:"bucketname"`
	TimeParseLayout    string `json:"timeparselayout"`
	NewRelicURL        string `json:"newrelicurl"`
	NewRelicAccountID  string `json:"newrelicaccountid"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if err := requireField("NewRelicLicenseKey", c.NewRelicLicenseKey); err != nil {
		return err
	}
	if err := requireField("NewRelicURL", c.NewRelicURL); err != nil {
		return err
	}
	if err := requireField("NewRelicAccountID", c.NewRelicAccountID); err != nil {
		return err
	}
	if err := requireField("BucketName", c.BucketName); err != nil {
		return err
	}
	if c.TimeParseLayout == "" {
		c.TimeParseLayout = "2006-01-02 15:04:05"
	}
	return nil
}

func requireField(name, value string) error {
	if value == "" {
		return fmt.Errorf("required field [%s] not found in config", name)
	}
	return nil
}
