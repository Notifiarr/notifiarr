package share //nolint:testpackage

import (
	"testing"
	"time"
)

func TestDeduperTake(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	duper := newDeduper(time.Hour)
	duper.now = func() time.Time { return now }

	line, exists := duper.take("boom")
	if !exists || line != "boom" {
		t.Fatalf("first take: got %q ok=%v, want boom true", line, exists)
	}

	line, exists = duper.take("boom")
	if exists || line != "" {
		t.Fatalf("second take during cooldown: got %q ok=%v, want empty false", line, exists)
	}

	_, _ = duper.take("boom")
	_, _ = duper.take("other")

	now = now.Add(time.Hour)

	line, exists = duper.take("boom")
	if !exists || line != "boom (repeated 2 times)" {
		t.Fatalf("after cooldown: got %q ok=%v, want annotated true", line, exists)
	}

	line, exists = duper.take("other")
	if !exists || line != "other" {
		t.Fatalf("other after cooldown with no extras: got %q ok=%v", line, exists)
	}
}

func TestDeduperPrune(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	duper := newDeduper(time.Hour)
	duper.now = func() time.Time { return now }

	_, _ = duper.take("old")
	now = now.Add(2 * time.Hour)
	_, _ = duper.take("new")

	if _, ok := duper.last["old"]; ok {
		t.Fatal("expected expired unused message to be pruned")
	}

	if _, ok := duper.last["new"]; !ok {
		t.Fatal("expected new message to remain")
	}
}
