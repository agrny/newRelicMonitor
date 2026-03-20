// Package newrelic
package newrelic

import (
	"fmt"
	"net/http"

	"new-relic-monitor/config"
)

type NewRelicClient struct {
	NewRelicLicenseKey string
	AppName            string
	APIURL             string
	Client             *http.Client
}

func NewClientWithConfig(options config.Config) (*NewRelicClient, error) {
	if options.NewRelicLicenseKey == "" {
		return nil, fmt.Errorf("new relic license key required")
	}

	return &NewRelicClient{
		NewRelicLicenseKey: options.NewRelicLicenseKey,
		Client:             &http.Client{},
	}, nil
}

func (nr *NewRelicClient) RecordCustomEvent(eventype string, data map[string]any) {



}
