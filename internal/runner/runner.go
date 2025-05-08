package runner

import (
	"bytes"
	"io"
	"os/exec"
	"sync"
	"bufio"
)

type Runner struct{}

func NewRunner() *Runner { return &Runner{} }

// startGoTest runs `go test -json` in the given pkg, optionally for a specific testName.
// Returns the exec.Cmd, a buffer containing output, and any startup error.
func (r *Runner) startGoTest(pkg string, testName string) (*exec.Cmd, *bytes.Buffer, error) {
	args := []string{"test", "-json"}
	if testName != "" {
		args = append(args, "-run", testName)
	}
	args = append(args, pkg)
	cmd := exec.Command("go", args...)
	cmd.Dir = findProjectRoot()

	out, err := cmd.CombinedOutput()
	buf := bytes.NewBuffer(out)
	if err != nil {
		buf.WriteString("\n[runner debug] cmd.Wait() error: " + err.Error() + "\n")
		return cmd, buf, err
	}
	return cmd, buf, nil
}

// Run executes go test -json and sends each output line to the channel.
func (r *Runner) Run(pkg string, testName string, out chan<- []byte) error {
	args := []string{"test", "-json"}
	if testName != "" {
		args = append(args, "-run", testName)
	}
	args = append(args, pkg)
	cmd := exec.Command("go", args...)
	cmd.Dir = findProjectRoot()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	stream := func(reader io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(bytes.TrimSpace(line)) > 0 {
				out <- append([]byte{}, line...)
			}
		}
		if err := scanner.Err(); err != nil {
			out <- []byte("[runner debug] scanner error: " + err.Error())
		}
	}
	wg.Add(2)
	go stream(stdout)
	go stream(stderr)
	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		// Log the error for debugging
		out <- []byte("[runner debug] cmd.Wait() error: " + err.Error())
	}
	return err
}


// findProjectRoot returns the absolute path to the project root directory.
func findProjectRoot() string {
	// Hardcoded for now; in a real implementation, this could be dynamic
	return "c:/Users/Remym/pythonProject/__personal-projects/go-sentinel"
}
