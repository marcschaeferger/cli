package notice

import (
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput runs fn with os.Stdout and os.Stderr redirected to pipes
// and returns what was written to each.
func captureOutput(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	origOut, origErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outW, errW
	defer func() { os.Stdout, os.Stderr = origOut, origErr }()

	outCh := make(chan string)
	errCh := make(chan string)
	go func() { b, _ := io.ReadAll(outR); outCh <- string(b) }()
	go func() { b, _ := io.ReadAll(errR); errCh <- string(b) }()

	fn()

	_ = outW.Close()
	_ = errW.Close()
	return <-outCh, <-errCh
}

func TestPrintBannerWritesToStderrNotStdout(t *testing.T) {
	stdout, stderr := captureOutput(t, func() {
		printBanner([]string{"Notice title", "notice body"})
	})

	if stdout != "" {
		t.Errorf("banner must not be written to stdout (breaks machine-readable output), got %q", stdout)
	}
	if !strings.Contains(stderr, "Notice title") || !strings.Contains(stderr, "notice body") {
		t.Errorf("banner should be written to stderr, got %q", stderr)
	}
}
