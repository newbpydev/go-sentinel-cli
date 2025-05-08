package runner

import (
	"os/exec"
)

// RunGoVersion runs 'go version' and returns its output and error. Useful for debugging Go environment.
func RunGoVersion() (string, error) {
	cmd := exec.Command("go", "version")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// RunGoEnv runs 'go env' and returns its output and error.
func RunGoEnv() (string, error) {
	cmd := exec.Command("go", "env")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// RunGoList runs 'go list <args>' and returns its output and error.
func RunGoList(args ...string) (string, error) {
	cmd := exec.Command("go", append([]string{"list"}, args...)...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// RunGoModTidy runs 'go mod tidy' and returns its output and error.
func RunGoModTidy() (string, error) {
	cmd := exec.Command("go", "mod", "tidy")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// RunGoFmt runs 'go fmt <packages>' and returns its output and error.
func RunGoFmt(pkgs ...string) (string, error) {
	cmd := exec.Command("go", append([]string{"fmt"}, pkgs...)...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
