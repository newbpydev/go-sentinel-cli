package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	// Create a command to run our app with the minimal test command
	cmd := exec.Command("./go-sentinel-cli", "minimaltest", "-v", "./simple_test.go")

	// Set the command to run in the current terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting command: %v\n", err)
		os.Exit(1)
	}

	// Create a channel to signal timeout
	timeout := time.After(30 * time.Second)

	// Create a channel to signal command completion
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Wait for either command completion or timeout
	select {
	case <-timeout:
		fmt.Println("\n\nTimeout reached (30 seconds). Killing process...")
		cmd.Process.Kill()
		os.Exit(0)
	case err := <-done:
		if err != nil {
			fmt.Printf("\n\nCommand finished with error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n\nCommand completed successfully")
		os.Exit(0)
	}
}
