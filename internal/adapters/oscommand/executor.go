package oscommand

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AntonioJCosta/nicksh/internal/core/ports"
)

// OSCommandExecutor implements the CommandExecutor interface using the operating system's shell.
type OSCommandExecutor struct{}

// NewOSCommandExecutor creates a new OSCommandExecutor.
func NewOSCommandExecutor() ports.CommandExecutor {
	return &OSCommandExecutor{}
}

// Execute runs the given pipeline string in a shell and returns its stdout, stderr, and any error.
// It attempts to use the system's default SHELL, falling back to common shells if not set.
func (e *OSCommandExecutor) Execute(shellName, pipeline string) (string, string, error) {
	shellExecPath := os.Getenv("SHELL")
	if shellExecPath == "" {
		// Fallback if SHELL environment variable is not set.
		switch shellName {
		case "bash":
			shellExecPath = "/bin/bash"
		case "zsh":
			shellExecPath = "/bin/zsh"
		default:
			// Default to sh if a specific known shell isn't requested or found.
			shellExecPath = "/bin/sh"
		}
	}

	cmd := exec.Command(shellExecPath, "-c", pipeline)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout := outBuf.String()
	stderr := errBuf.String()

	if err != nil {
		// Include stderr in the error message for better diagnostics.
		return stdout, stderr, fmt.Errorf("executing pipeline with shell '%s': %w. Stderr: %s", shellExecPath, err, strings.TrimSpace(stderr))
	}
	return stdout, stderr, nil
}
