package share //nolint:testpackage

import (
	"testing"
	"time"
)

func TestShareKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
	}{
		{in: "boom", want: "boom"},
		{in: "{trace:abc-123} boom", want: "boom"},
		{in: "{trace:abc-123} Unable to watch file: x", want: "Unable to watch file: x"},
		{in: "{trace:no-close boom", want: "{trace:no-close boom"},
	}

	for _, test := range tests {
		if got := shareKey(test.in); got != test.want {
			t.Fatalf("shareKey(%q) = %q, want %q", test.in, got, test.want)
		}
	}
}

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

func TestDeduperRepeatedOnce(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	duper := newDeduper(time.Hour)
	duper.now = func() time.Time { return now }

	_, _ = duper.take("boom")
	_, _ = duper.take("boom")
	now = now.Add(time.Hour)

	line, exists := duper.take("boom")
	if !exists || line != "boom (repeated 1 time)" {
		t.Fatalf("got %q ok=%v, want singular repeated 1 time", line, exists)
	}
}

func TestDeduperTraceKey(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	duper := newDeduper(time.Hour)
	duper.now = func() time.Time { return now }

	line, exists := duper.take("{trace:req-1} boom")
	if !exists || line != "{trace:req-1} boom" {
		t.Fatalf("first traced take: got %q ok=%v", line, exists)
	}

	line, exists = duper.take("{trace:req-2} boom")
	if exists || line != "" {
		t.Fatalf("second traced take should dedupe on untraced key, got %q ok=%v", line, exists)
	}

	now = now.Add(time.Hour)

	line, exists = duper.take("{trace:req-3} boom")
	if !exists || line != "{trace:req-3} boom (repeated 1 time)" {
		t.Fatalf("after cooldown: got %q ok=%v", line, exists)
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

func TestDeduperPruneSkipped(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	duper := newDeduper(time.Hour)
	duper.now = func() time.Time { return now }

	_, _ = duper.take("once")
	_, _ = duper.take("once") // skipped=1, would previously never prune

	now = now.Add(time.Hour)
	_, _ = duper.take("other")

	if _, ok := duper.last["once"]; !ok {
		t.Fatal("skip count should be retained for one extra cooldown")
	}

	now = now.Add(time.Hour)
	_, _ = duper.take("other")

	if _, ok := duper.last["once"]; ok {
		t.Fatal("expected skipped-only message to be pruned after 2 cooldowns")
	}

	if duper.skipped["once"] != 0 {
		t.Fatal("expected skipped count to be dropped with the key")
	}
}
