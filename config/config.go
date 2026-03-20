package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	AWSKeySecret       string
	AWSKeyID           string
	NewRelicLicenseKey string
	BucketName         string
	TimeParseLayout    string
}

func (c *Config) ReadConfigFile(path string) {
	configOptions := Config{}
	data, _ := os.ReadFile(path)
	json.Unmarshal(data, &configOptions)

	c.AWSKeyID = configOptions.AWSKeyID
	c.AWSKeySecret = configOptions.AWSKeySecret
	c.BucketName = configOptions.BucketName
	c.TimeParseLayout = configOptions.TimeParseLayout
}
