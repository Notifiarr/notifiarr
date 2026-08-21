package filewatch //nolint:testpackage

import (
	"fmt"
	"testing"
)

func TestIsQuietSetupErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "ignored log", err: fmt.Errorf("%w: %s", ErrIgnoredLog, "Notifiarr.log"), want: true},
		{name: "disabled", err: fmt.Errorf("%w: %s", ErrDisabled, "foo.log"), want: true},
		{
			name: "invalid regexp",
			err:  fmt.Errorf("%w: no regexp match provided, ignored: %s", ErrInvalidRegexp, "foo"),
			want: false,
		},
		{name: "other", err: fmt.Errorf("watching file %s: no such file", "foo.log"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := isQuietSetupErr(test.err); got != test.want {
				t.Fatalf("isQuietSetupErr(%v) = %v, want %v", test.err, got, test.want)
			}
		})
	}
}
