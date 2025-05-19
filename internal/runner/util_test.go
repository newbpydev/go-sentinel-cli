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
		{name: "standard package", args: []string{"-f", "{{.ImportPath}}", "fmt"}},
		{name: "dot package", args: []string{"-f", "{{.Name}}", "."}},
		{name: "all packages", args: []string{"-e", "all"}},
		{name: "test package", args: []string{"-test", "."}},
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
		{name: "command injection", args: []string{"; ls"}},
		{name: "pipe", args: []string{"| echo"}},
		{name: "redirect", args: []string{"> file"}},
		{name: "backticks", args: []string{"`pwd`"}},
		{name: "dollar sign", args: []string{"$PATH"}},
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
		{name: "current package", paths: []string{"."}, wantErr: false},
		{name: "specific package", paths: []string{"./internal/runner"}, wantErr: false},
		{name: "multiple packages", paths: []string{"./internal/runner", "./internal/parser"}, wantErr: false},
		{name: "invalid path", paths: []string{"../../../etc/passwd"}, wantErr: true},
		{name: "command injection", paths: []string{"; ls"}, wantErr: true},
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
