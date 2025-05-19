package runner

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRunGoTestJSONInCorrectPkg(t *testing.T) {
	r := NewRunner()
	pkg := "./internal/runner/testdata/passonly"
	cmd, out, err := r.startGoTest(pkg, "")
	if err != nil {
		t.Fatalf("failed to start go test: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("go test command failed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, `"Action":"pass"`) {
		t.Errorf("expected JSON output with pass action, got: %q", output)
	}
}

func TestCaptureStdoutStderrAndHandleErrors(t *testing.T) {
	r := NewRunner()
	pkg := "./internal/runner/testdata/passonly"
	cmd, out, err := r.startGoTest(pkg, "")
	if err != nil {
		t.Fatalf("failed to start go test: %v", err)
	}

	// Wait for the command to complete
	err = cmd.Wait()
	if err != nil {
		t.Fatalf("go test command failed: %v", err)
	}

	// Check if we got any output
	if out.Len() == 0 {
		t.Error("expected non-empty stdout from go test, but got empty output")
	}
}

func TestHandleNonJSONOutput(t *testing.T) {
	r := NewRunner()
	pkg := "./internal/runner/testdata/badbuild"
	cmd, out, err := r.startGoTest(pkg, "")
	if err != nil {
		t.Fatalf("failed to start go test: %v", err)
	}

	// Wait for the command to complete
	err = cmd.Wait()
	if err == nil {
		t.Error("expected error for build failure, but got none")
	}

	// Check if we got any output
	output := out.Bytes()
	if len(output) == 0 {
		t.Fatal("expected output from go test for build failure, but got none")
	}

	// Check for build error in output
	if !bytes.Contains(bytes.ToLower(output), []byte("build")) {
		t.Errorf("expected build error in output, got: %q", output)
	}
}

func TestPipeStdoutStderrForRealtimeOutput(t *testing.T) {
	r := NewRunner()
	pkg := "./internal/runner/testdata/passonly"
	cmd, out, err := r.startGoTest(pkg, "")
	if err != nil {
		t.Fatalf("failed to start go test: %v", err)
	}
	done := make(chan struct{})
	go func() {
		if err := cmd.Wait(); err != nil {
			t.Errorf("go test command failed: %v", err)
		}
		close(done)
	}()
	select {
	case <-done:
		if out.Len() == 0 {
			t.Errorf("expected output from go test, got: %q", out.Bytes())
		}
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for go test output")
	}
}

func TestDebugGoVersionUtil(t *testing.T) {
	_, err := RunGoVersion()
	if err != nil {
		t.Errorf("go version error: %v", err)
	}
}

func TestIntegrationWithGoroutinePipeline(t *testing.T) {
	r := NewRunner()
	pkg := "./internal/runner/testdata/passonly"
	ch := make(chan []byte, 8)
	go func() {
		_ = r.Run(pkg, "", ch)
		close(ch)
	}()
	var got bool
	for line := range ch {
		if strings.Contains(string(line), `"Action":"pass"`) {
			got = true
		}
	}
	if !got {
		t.Errorf("expected pass action in pipeline output")
	}
}
