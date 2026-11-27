package misc

import (
	"errors"
	"testing"
)

func TestStopWithError(t *testing.T) {
	if err := StopError(); err != nil {
		t.Fatalf("unexpected error before the TUI was stopped: %v", err)
	}

	first := errors.New("first")
	StopWithError(first)
	StopWithError(errors.New("second"))

	// The error that stopped the TUI is the one to report
	if err := StopError(); err != first {
		t.Fatalf("got error %v, want %v", err, first)
	}
}
