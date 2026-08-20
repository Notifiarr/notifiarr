package services_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/cnfg"
)

func TestServiceConfigValidate(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name    string
		cfg     services.ServiceConfig
		wantErr error
		after   func(*testing.T, services.ServiceConfig)
	}{
		{
			name:    "missing name",
			cfg:     services.ServiceConfig{Type: services.CheckTCP, Value: "127.0.0.1:80"},
			wantErr: services.ErrNoName,
		},
		{
			name:    "missing value",
			cfg:     services.ServiceConfig{Name: "empty", Type: services.CheckHTTP},
			wantErr: services.ErrNoCheck,
		},
		{
			name:    "invalid type",
			cfg:     services.ServiceConfig{Name: "ftp", Type: "ftp", Value: "x"},
			wantErr: services.ErrInvalidType,
		},
		{
			name:    "tcp missing port",
			cfg:     services.ServiceConfig{Name: "tcp", Type: services.CheckTCP, Value: "localhost"},
			wantErr: services.ErrBadTCP,
		},
		{
			name: "http defaults and clamps",
			cfg: services.ServiceConfig{
				Name:     "http",
				Type:     services.CheckHTTP,
				Value:    "http://example.com",
				Timeout:  cnfg.Duration{Duration: 50 * time.Millisecond},
				Interval: cnfg.Duration{Duration: time.Second},
			},
			after: func(t *testing.T, cfg services.ServiceConfig) {
				t.Helper()
				assert.Equal(t, "200", cfg.Expect)
				assert.Equal(t, services.MinimumTimeout, cfg.Timeout.Duration)
				assert.Equal(t, services.MinimumCheckInterval, cfg.Interval.Duration)
			},
		},
		{
			name: "http default timeout when unset",
			cfg: services.ServiceConfig{
				Name:  "http-timeout",
				Type:  services.CheckHTTP,
				Value: "http://example.com",
			},
			after: func(t *testing.T, cfg services.ServiceConfig) {
				t.Helper()
				assert.Equal(t, services.DefaultTimeout, cfg.Timeout.Duration)
			},
		},
		{
			name: "http ssl expect is accepted",
			cfg: services.ServiceConfig{
				Name:   "https",
				Type:   services.CheckHTTP,
				Value:  "https://example.com",
				Expect: "200,SSL",
			},
		},
		{
			name: "tcp ok",
			cfg: services.ServiceConfig{
				Name:  "tcp-ok",
				Type:  services.CheckTCP,
				Value: "127.0.0.1:3306",
			},
		},
		{
			name: "process missing value is no-check",
			cfg: services.ServiceConfig{
				Name: "proc",
				Type: services.CheckPROC,
			},
			wantErr: services.ErrNoCheck,
		},
		{
			name: "process invalid expect",
			cfg: services.ServiceConfig{
				Name:   "proc-bad",
				Type:   services.CheckPROC,
				Value:  "nginx",
				Expect: "nope",
			},
			wantErr: services.ErrProcExpect,
		},
		{
			name: "process running with count",
			cfg: services.ServiceConfig{
				Name:   "proc-count-running",
				Type:   services.CheckPROC,
				Value:  "nginx",
				Expect: "running,count:1:2",
			},
			wantErr: services.ErrCountZero,
		},
		{
			name: "process invalid min count",
			cfg: services.ServiceConfig{
				Name:   "proc-min",
				Type:   services.CheckPROC,
				Value:  "nginx",
				Expect: "count:abc",
			},
		},
		{
			name: "process invalid max count",
			cfg: services.ServiceConfig{
				Name:   "proc-max",
				Type:   services.CheckPROC,
				Value:  "nginx",
				Expect: "count:1:xyz",
			},
		},
		{
			name: "process invalid regex",
			cfg: services.ServiceConfig{
				Name:  "proc-re",
				Type:  services.CheckPROC,
				Value: "/(unclosed/",
			},
		},
		{
			name: "process ok with regex and counts",
			cfg: services.ServiceConfig{
				Name:   "proc-ok",
				Type:   services.CheckPROC,
				Value:  "/nginx|caddy/",
				Expect: "count:1:3,restart",
			},
		},
		{
			name: "ping missing value is no-check",
			cfg: services.ServiceConfig{
				Name: "ping",
				Type: services.CheckPING,
			},
			wantErr: services.ErrNoCheck,
		},
		{
			name: "ping bad expect shape",
			cfg: services.ServiceConfig{
				Name:   "ping-shape",
				Type:   services.CheckPING,
				Value:  "127.0.0.1",
				Expect: "3:2",
			},
			wantErr: services.ErrPingExpect,
		},
		{
			name: "ping zero interval",
			cfg: services.ServiceConfig{
				Name:   "ping-zero",
				Type:   services.CheckICMP,
				Value:  "127.0.0.1",
				Expect: "3:2:0",
			},
			wantErr: services.ErrPingExpect,
		},
		{
			name: "ping min greater than count",
			cfg: services.ServiceConfig{
				Name:   "ping-min",
				Type:   services.CheckPING,
				Value:  "127.0.0.1",
				Expect: "2:3:100",
			},
			wantErr: services.ErrPingExpect,
		},
		{
			name: "ping invalid count",
			cfg: services.ServiceConfig{
				Name:   "ping-count",
				Type:   services.CheckPING,
				Value:  "127.0.0.1",
				Expect: "x:1:100",
			},
		},
		{
			name: "ping invalid min",
			cfg: services.ServiceConfig{
				Name:   "ping-min-parse",
				Type:   services.CheckPING,
				Value:  "127.0.0.1",
				Expect: "1:x:100",
			},
		},
		{
			name: "ping invalid interval",
			cfg: services.ServiceConfig{
				Name:   "ping-int-parse",
				Type:   services.CheckPING,
				Value:  "127.0.0.1",
				Expect: "1:1:x",
			},
		},
		{
			name: "ping ok",
			cfg: services.ServiceConfig{
				Name:   "ping-ok",
				Type:   services.CheckPING,
				Value:  "127.0.0.1",
				Expect: "3:2:100",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cfg := test.cfg
			err := cfg.Validate()
			wantErr := test.wantErr

			if test.name == "process ok with regex and counts" && runtime.GOOS == "freebsd" {
				wantErr = services.ErrBSDRestart
			}

			if wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, wantErr)

				return
			}

			parseErrors := map[string]struct{}{
				"process invalid min count": {},
				"process invalid max count": {},
				"process invalid regex":     {},
				"ping invalid count":        {},
				"ping invalid min":          {},
				"ping invalid interval":     {},
			}
			if _, isParseErr := parseErrors[test.name]; isParseErr {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error())

				return
			}

			require.NoError(t, err)

			if test.after != nil {
				test.after(t, cfg)
			}
		})
	}
}
