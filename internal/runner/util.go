// Package runner provides functionality for executing Go tests and processing their results.
// It includes tools for running tests, parsing output, and handling test events.
package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	// Handle template flags separately
	var cmdArgs []string
	cmdArgs = append(cmdArgs, "list")
	for _, arg := range args {
		if strings.HasPrefix(arg, "-f=") || strings.HasPrefix(arg, "-format=") {
			cmdArgs = append(cmdArgs, arg)
		} else if arg == "-f" || arg == "-format" {
			// Skip the flag and its value for now
			continue
		} else if strings.HasPrefix(arg, "{{") {
			// This is a template value, append it with the previous -f flag
			cmdArgs = append(cmdArgs, "-f", arg)
		} else {
			cmdArgs = append(cmdArgs, arg)
		}
	}

	cmd := exec.Command("go", cmdArgs...) // #nosec G204
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
func RunGoFmt(paths ...string) (string, error) {
	// Validate package paths to prevent command injection
	for _, path := range paths {
		if !isValidPackagePath(path) {
			return "", fmt.Errorf("invalid package path: %s", path)
		}
	}

	// Convert relative paths to absolute paths
	var absPaths []string
	for _, path := range paths {
		if strings.HasPrefix(path, ".") {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("failed to resolve absolute path for %s: %v", path, err)
			}
			absPaths = append(absPaths, absPath)
		} else {
			absPaths = append(absPaths, path)
		}
	}

	cmd := exec.Command("go", append([]string{"fmt"}, absPaths...)...) // #nosec G204
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// isValidGoCommandArg checks if an argument is safe to pass to go commands.
func isValidGoCommandArg(arg string) bool {
	// Allow common go command flags
	if strings.HasPrefix(arg, "-") {
		// Allow template flags with their values
		if strings.HasPrefix(arg, "-f=") || strings.HasPrefix(arg, "-format=") {
			return true
		}
		validFlags := map[string]bool{
			"-f":      true,
			"-format": true,
			"-e":      true,
			"-test":   true,
			"-json":   true,
			"-short":  true,
			"-v":      true,
		}
		return validFlags[arg]
	}

	// Allow package paths and patterns
	return !strings.ContainsAny(arg, ";|><$`\\")
}

// isValidPackagePath checks if a package path is safe to pass to go commands.
func isValidPackagePath(path string) bool {
	// Allow relative paths starting with .
	if strings.HasPrefix(path, ".") {
		return !strings.ContainsAny(path, ";|><$`\\")
	}
	// Allow absolute paths within the workspace
	if strings.HasPrefix(path, "/") {
		wd, err := os.Getwd()
		if err != nil {
			return false
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			return false
		}
		return strings.HasPrefix(absPath, wd)
	}
	return false
}
