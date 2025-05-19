package runner

import (
	"strings"
	"testing"
)

func TestRunGoVersion_Output(t *testing.T) {
	out, err := RunGoVersion()
	if err != nil {
		t.Fatalf("RunGoVersion failed: %v", err)
	}
	if !strings.Contains(out, "go version") {
		t.Errorf("Expected output to contain 'go version', got %q", out)
	}
}

func TestRunGoEnv_Output(t *testing.T) {
	out, err := RunGoEnv()
	if err != nil {
		t.Fatalf("RunGoEnv failed: %v", err)
	}
	if !strings.Contains(out, "GOPATH") {
		t.Errorf("Expected output to contain 'GOPATH', got %q", out)
	}
}

func TestRunGoList_ValidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"standard package", {"-f", "{{.ImportPath}}", "fmt"}},
		{"dot package", {"-f", "{{.Name}}", "."}},
		{"all packages", {"-e", "all"}},
		{"test package", {"-test", "."}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := RunGoList(tc.args...)
			if err != nil {
				t.Errorf("RunGoList(%v) failed: %v", tc.args, err)
			}
			if out == "" {
				t.Errorf("RunGoList(%v) returned empty output", tc.args)
			}
		})
	}
}

func TestRunGoList_InvalidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"command injection", {"; ls"}},
		{"pipe", {"| echo"}},
		{"redirect", {"> file"}},
		{"backticks", {"`pwd`"}},
		{"dollar sign", {"$PATH"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := RunGoList(tc.args...)
			if err == nil {
				t.Errorf("RunGoList(%v) should have failed but didn't", tc.args)
			}
		})
	}
}

func TestRunGoModTidy_Output(t *testing.T) {
	out, err := RunGoModTidy()
	if err != nil {
		t.Fatalf("RunGoModTidy failed: %v", err)
	}
	// go mod tidy might not output anything if modules are already tidy
	// so we just check that the command executed successfully
	t.Logf("RunGoModTidy output: %q", out)
}

func TestRunGoFmt_ValidPaths(t *testing.T) {
	tests := []struct {
		name    string
		paths   []string
		wantErr bool
	}{
		{"current package", {"."}, false},
		{"specific package", {"./internal/runner"}, false},
		{"multiple packages", {"./internal/runner", "./internal/parser"}, false},
		{"invalid path", {"../../../etc/passwd"}, true},
		{"command injection", {"; ls"}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := RunGoFmt(tc.paths...)
			if tc.wantErr {
				if err == nil {
					t.Errorf("RunGoFmt(%v) should have failed but didn't", tc.paths)
				}
			} else {
				if err != nil {
					t.Errorf("RunGoFmt(%v) failed: %v", tc.paths, err)
				}
			}
		})
	}
}
