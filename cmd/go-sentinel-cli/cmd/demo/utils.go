package demo

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// printJSON pretty-prints an object as JSON
func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "    ", "  ")
	if err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// formatBytes formats bytes as human-readable string
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatUint(bytes, 10) + " B"
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// isColorTerminal detects if the terminal supports colors
func isColorTerminal() bool {
	// Check for NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check for FORCE_COLOR environment variable
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check for TTY
	fileInfo, _ := os.Stdout.Stat()
	if (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}

	// Check for specific environment variables
	term := os.Getenv("TERM")
	if term == "xterm" || term == "xterm-256color" || term == "screen" || term == "screen-256color" {
		return true
	}

	return false
}
