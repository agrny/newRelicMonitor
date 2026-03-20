package newrelic

import (
	"os"

	"new-relic-monitor/config"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func NewApplicationWithConfig(options config.Config) (*newrelic.Application, error) {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("S3 Backup Monitor"),
		newrelic.ConfigLicense(options.NewRelicLicenseKey),
		newrelic.ConfigDebugLogger(os.Stdout),
	)
	if err != nil {
		return nil, err
	}
	return app, nil
}
