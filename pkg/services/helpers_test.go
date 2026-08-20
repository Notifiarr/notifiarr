package services_test

import (
	"os"
	"testing"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"golift.io/cnfg"
)

func TestMain(m *testing.M) {
	mnd.Log = logs.Log
	os.Exit(m.Run())
}

func resultsByName(results []*services.CheckResult) map[string]*services.CheckResult {
	out := make(map[string]*services.CheckResult, len(results))
	for _, result := range results {
		out[result.Name] = result
	}

	return out
}

func extraConfig(name string, interval time.Duration) apps.ExtraConfig {
	return apps.ExtraConfig{
		Name:     name,
		Timeout:  cnfg.Duration{Duration: time.Second},
		Interval: cnfg.Duration{Duration: interval},
	}
}

func starrApp(name, appURL, apiKey string, interval time.Duration) apps.StarrApp {
	return apps.StarrApp{
		URL: appURL, APIKey: apiKey,
		ExtraConfig: extraConfig(name, interval),
	}
}
