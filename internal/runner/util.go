// Package runner provides functionality for executing Go tests and processing their results.
// It includes tools for running tests, parsing output, and handling test events.
package runner

import (
	"fmt"
	"os/exec"
	"strings"
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
// It validates the arguments to prevent command injection.
// G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
// Args are validated by isValidGoCommandArg before use
func RunGoList(args ...string) (string, error) {
	// Validate arguments to prevent command injection
	for _, arg := range args {
		if !isValidGoCommandArg(arg) {
			return "", fmt.Errorf("invalid argument for go list: %s", arg)
		}
	}

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
// It validates the package paths to prevent command injection.
// G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
// Args are validated by isValidPackagePath before use
func RunGoFmt(pkgs ...string) (string, error) {
	// Validate package paths to prevent command injection
	for _, pkg := range pkgs {
		if !isValidPackagePath(pkg) {
			return "", fmt.Errorf("invalid package path for go fmt: %s", pkg)
		}
	}

	cmd := exec.Command("go", append([]string{"fmt"}, pkgs...)...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// isValidGoCommandArg checks if a string is a valid argument for Go commands.
// This helps prevent command injection by rejecting suspicious arguments.
func isValidGoCommandArg(arg string) bool {
	// Reject arguments that might be used for command injection
	if strings.Contains(arg, ";") || strings.Contains(arg, "&") ||
		strings.Contains(arg, "|") || strings.Contains(arg, ">") ||
		strings.Contains(arg, "<") || strings.Contains(arg, "`") ||
		strings.HasPrefix(arg, "-") || strings.Contains(arg, "$(") ||
		strings.Contains(arg, "${") {
		return false
	}

	// Additional validation for Go package/module paths
	return true
}

// isValidPackagePath checks if a string is a valid Go package path.
// This helps prevent command injection in package paths.
func isValidPackagePath(path string) bool {
	// Basic validation for Go package paths
	if path == "" {
		return false
	}

	// Check for suspicious characters that might enable command injection
	return isValidGoCommandArg(path)
}
