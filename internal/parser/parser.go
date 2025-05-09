package parser

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/newbpydev/go-sentinel/internal/ui"
)

type goTestEvent struct {
	Time    string `json:"Time"`
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
	Output  string `json:"Output"`
}

var fileLineRe = regexp.MustCompile(`(?m)^\s*(.+\.go):(\d+):`)

// ParseTestResults runs 'go test -json ./...' and parses output into []ui.TestResult
func ParseTestResults() ([]ui.TestResult, error) {
	cmd := exec.Command("go", "test", "-json", "./...")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	// Suppress stderr from leaking to terminal
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	defer cmd.Wait()

	scanner := bufio.NewScanner(stdout)
	var results []ui.TestResult
	failures := map[string]*ui.TestResult{}
	for scanner.Scan() {
		line := scanner.Bytes()
		var ev goTestEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		if ev.Action == "fail" && ev.Test != "" {
			// Mark as failed test
			tr := ui.TestResult{
				Package: ev.Package,
				Passed:  false,
				Summary: "FAIL " + ev.Package + "/" + ev.Test,
			}
			failures[ev.Package+"/"+ev.Test] = &tr
			results = append(results, tr)
		}
		if ev.Action == "output" && ev.Test != "" {
			if tr, ok := failures[ev.Package+"/"+ev.Test]; ok {
				msg := strings.TrimSpace(ev.Output)
				if msg != "" {
					tr.Message += msg + "\n"
					// Try to extract file/line
					if tr.File == "" {
						m := fileLineRe.FindStringSubmatch(msg)
						if len(m) == 3 {
							tr.File = m[1]
							tr.Line = atoiSafe(m[2])
						}
					}
				}
			}
		}
		if ev.Action == "pass" && ev.Test != "" {
			// Mark as passed test
			results = append(results, ui.TestResult{
				Package: ev.Package,
				Passed:  true,
				Summary: "ok   " + ev.Package + "/" + ev.Test,
			})
		}
	}
	return results, nil
}

func atoiSafe(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
