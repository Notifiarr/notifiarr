package services

import (
	"testing"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"golift.io/cnfg"
)

func TestCollectMySQLAppsPreservesHostnames(t *testing.T) {
	t.Parallel()

	mysql := []snapshot.MySQLConfig{
		{Name: "privatebin", Host: "privatebin-mariadb", Timeout: cnfg.Duration{Duration: time.Second}},
		{Name: "loopback", Host: "@tcp(127.0.0.1:3306)", Timeout: cnfg.Duration{Duration: time.Second}},
	}

	svcs := collectMySQLApps(nil, mysql)
	if len(svcs) != 2 {
		t.Fatalf("expected 2 service checks, got %d", len(svcs))
	}

	if got, want := svcs[0].Value, "privatebin-mariadb:3306"; got != want {
		t.Fatalf("privatebin host name mismatch: got %q, want %q", got, want)
	}

	if got, want := svcs[1].Value, "127.0.0.1:3306"; got != want {
		t.Fatalf("tcp host value mismatch: got %q, want %q", got, want)
	}
}
