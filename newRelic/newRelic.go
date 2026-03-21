// Package newrelic
package newrelic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type NewRelicClient struct {
	NewRelicLicenseKey string
	AppName            string
	APIURL             string
	Client             *http.Client
}

func NewClientWithConfig(options Config) (*NewRelicClient, error) {
	if options.NewRelicLicenseKey == "" {
		return nil, fmt.Errorf("newreliclicensekey field required")
	}

	return &NewRelicClient{
		NewRelicLicenseKey: options.NewRelicLicenseKey,
		Client:             &http.Client{},
		APIURL:             fmt.Sprintf("%s/%s/events", options.NewRelicURL, options.NewRelicAccountID),
	}, nil
}

func (nr *NewRelicClient) RecordCustomEvent(eventype string, data map[string]any) error {
	dataMarshalled, err := json.Marshal(data)
	if err != nil {
		return err
	}
	body := bytes.NewReader(dataMarshalled)

	req, err := http.NewRequest(http.MethodPost, nr.APIURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Api-Key", nr.NewRelicLicenseKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := nr.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		fmt.Println("POST SUCCESS STATUS OK")
	}

	rawResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("\nCODE: %d RESPONSE: %s\n", resp.StatusCode, string(rawResponse))

	return nil
}
