package runner

import (
	"bytes"
	"testing"
	"time"
	"strings"
)

func TestRunGoTestJSONInCorrectPkg(t *testing.T) {
	r := NewRunner()
	pkg := "./internal/runner/testdata/passonly"
	cmd, out, err := r.startGoTest(pkg, "")
	if err != nil {
		t.Fatalf("failed to start go test: %v", err)
	}
	cmd.Wait()
	output := string(out.Bytes())
if !strings.Contains(output, `"Action":"pass"`) {
	t.Errorf("expected JSON output with pass action, got: %q", output)
}
}

func TestCaptureStdoutStderrAndHandleErrors(t *testing.T) {
	r := NewRunner()
	pkg := "./internal/runner/testdata/passonly"
	_, out, err := r.startGoTest(pkg, "")
	if err != nil {
		t.Fatalf("failed to start go test: %v", err)
	}
	if out.Len() == 0 {
		t.Errorf("expected non-empty stdout from go test, got: %q", out.Bytes())
	}
}

func TestHandleNonJSONOutput(t *testing.T) {
	r := NewRunner()
	pkg := "./internal/runner/testdata/badbuild"
	_, out, err := r.startGoTest(pkg, "")
	if err == nil {
		t.Error("expected error for build failure")
	}
	if out == nil || out.Len() == 0 {
		t.Fatalf("expected output from go test for build failure, got: %q", out.Bytes())
	}
	if !bytes.Contains(bytes.ToLower(out.Bytes()), []byte("build")) {
		t.Errorf("expected build error output, got: %q", out.Bytes())
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
		cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
		if out.Len() == 0 {
			t.Errorf("expected output from go test, got: %q", out.Bytes())
		}
	case <-time.After(2 * time.Second):
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
