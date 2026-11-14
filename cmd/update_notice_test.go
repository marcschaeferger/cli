package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	versionpkg "github.com/fosrl/cli/internal/version"
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

// The update banner is printed from a background goroutine while any command
// runs, including ones whose stdout is consumed by programs (e.g. sshd reading
// `auth-daemon principals`, or `status --json`). It must stay off stdout.
func TestUpdateAvailableNoticeWritesToStderrNotStdout(t *testing.T) {
	stdout, stderr := captureOutput(t, func() {
		showUpdateAvailable(&versionpkg.GitHubRelease{TagName: "9.9.9"})
	})

	if stdout != "" {
		t.Errorf("update notice must not be written to stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "A new version is available: 9.9.9") {
		t.Errorf("update notice should be written to stderr, got %q", stderr)
	}
	if !strings.Contains(stderr, "pangolin update") {
		t.Errorf("update notice should tell the user how to update, got %q", stderr)
	}
}
