package history

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/history"
)

// toUserFriendlyPath converts an absolute path to a ~/-based path if it's under the user's home directory.
// If the home directory cannot be determined or the path is not under home, it returns the original path.
func toUserFriendlyPath(absPath string) string {
	usr, err := user.Current()
	if err != nil {
		return absPath // Fallback if user/home directory cannot be determined
	}
	homeDir := usr.HomeDir

	if !strings.HasPrefix(absPath, homeDir) {
		return absPath // Path is not under home directory
	}

	if absPath == homeDir {
		return "~"
	}

	relPath, err := filepath.Rel(homeDir, absPath)
	if err != nil {
		return absPath // Fallback in case of an unexpected error with Rel
	}
	return filepath.Join("~", relPath)
}

// findUserHistoryFile attempts to find a shell history file by checking common locations and environment variables.
// It no longer takes shellExecutablePath as an argument.
func findUserHistoryFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("getting current user: %w", err)
	}
	homeDir := usr.HomeDir

	// 1. Check HISTFILE environment variable (common for Zsh, respected by some other tools)
	if histFileEnvVal := os.Getenv("HISTFILE"); histFileEnvVal != "" {
		pathToCheck := histFileEnvVal
		if !filepath.IsAbs(pathToCheck) {
			// Resolve relative to home directory if not absolute
			pathToCheck = filepath.Join(homeDir, pathToCheck)
		}
		if _, err := os.Stat(pathToCheck); err == nil {
			return pathToCheck, nil
		}
		// If HISTFILE is set but file doesn't exist, we can optionally log a warning
		// fmt.Fprintf(os.Stderr, "Warning: HISTFILE environment variable is set to '%s' but the file was not found.\n", histFileEnvVal)
	}

	// 2. Check a list of common default history file paths
	// Order can be significant if a user somehow has multiple (e.g. switched shells).
	potentialPaths := []string{
		filepath.Join(homeDir, ".zsh_history"),  // Common for Zsh
		filepath.Join(homeDir, ".bash_history"), // Common for Bash
		// Add other common paths here if desired, e.g.:
		// filepath.Join(homeDir, ".local", "share", "fish", "fish_history"), // Common for Fish (XDG)
		// filepath.Join(homeDir, ".config", "fish", "fish_history"),       // Older Fish
		// filepath.Join(homeDir, ".history"), // A generic fallback some might use
	}

	for _, p := range potentialPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil // Found a history file
		}
	}

	return "", fmt.Errorf("could not automatically find a common shell history file. Please ensure your history file is in a standard location (e.g., ~/.bash_history, ~/.zsh_history) or set the HISTFILE environment variable")
}

// parsePipelineOutput is a helper to parse the output of the shell pipeline.
func parsePipelineOutput(output string) ([]history.CommandFrequency, error) {
	frequencies := []history.CommandFrequency{}
	lines := strings.SplitSeq(output, "\n") // Changed from strings.SplitSeq for broader compatibility if needed
	for line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		parts := strings.Fields(trimmedLine)
		if len(parts) < 2 {
			continue // Skip lines that don't have at least count and command
		}
		count, err := strconv.Atoi(parts[0])
		if err != nil {
			// Log or skip malformed lines
			// fmt.Fprintf(os.Stderr, "Warning: could not parse count from history line: %s\n", trimmedLine)
			continue
		}
		command := strings.Join(parts[1:], " ")
		frequencies = append(frequencies, history.CommandFrequency{Command: command, Count: count})
	}
	return frequencies, nil
}

// determineScanCount determines how many history entries to scan.
func determineScanCount(fcHistoryScanLimit int) (int, error) {
	if fcHistoryScanLimit > 0 { // User-defined limit takes precedence
		return fcHistoryScanLimit, nil
	}
	histSizeStr := os.Getenv("HISTSIZE")
	if histSizeStr != "" {
		if histSize, err := strconv.Atoi(histSizeStr); err == nil && histSize > 0 {
			return histSize, nil
		}
		// Optionally log error if HISTSIZE is set but not a valid number
	}
	return 500, nil // Default scan count
}

// getHistoryFrequencies is a method on HistoryProvider (assuming HistoryProvider struct is defined elsewhere).
// It uses p.HistoryFile, which should be populated by calling findUserHistoryFile() during provider initialization.
func (p *HistoryProvider) getHistoryFrequencies(scanLimit, outputLimit int) ([]history.CommandFrequency, error) {
	if p.HistoryFile == "" {
		return nil, fmt.Errorf("history file path is not set in HistoryProvider")
	}
	scanCountVal, _ := determineScanCount(scanLimit) // Error from determineScanCount is ignored as it provides a default
	pipeline, err := buildShellPipeline(p.HistoryFile, strconv.Itoa(scanCountVal), outputLimit)
	if err != nil {
		return nil, fmt.Errorf("building shell pipeline: %w", err)
	}

	// Use the injected command executor (p.cmdExecutor) and shell (p.Shell)
	stdout, stderrOutput, err := p.cmdExecutor.Execute(p.Shell, pipeline)
	if err != nil {
		// The error from OSCommandExecutor.Execute might already include stderr.
		// Consider how to best present this error.
		errMsg := fmt.Sprintf("executing shell pipeline: %v", err)
		if stderrOutput != "" {
			errMsg = fmt.Sprintf("%s. Stderr: %s", errMsg, stderrOutput)
		}
		if stdout != "" {
			return nil, fmt.Errorf("%s. Stdout: %s", errMsg, stdout)
		}
		return nil, errors.New("history file path is not set in HistoryProvider")

	}
	// if stderrOutput != "" { // Log non-fatal stderr if necessary
	// 	fmt.Fprintf(os.Stderr, "Shell pipeline stderr: %s\n", stderrOutput)
	// }

	return parsePipelineOutput(stdout)
}

// buildShellPipeline constructs the shell command pipeline for bash/zsh.
func buildShellPipeline(historyFilePath, historyScanCountStr string, outputLimit int) (string, error) {
	if _, err := os.Stat(historyFilePath); os.IsNotExist(err) {
		// Use toUserFriendlyPath for displaying the path in the error message
		return "", fmt.Errorf("history file does not exist: %s", toUserFriendlyPath(historyFilePath))
	}
	// Ensure outputLimit is positive
	if outputLimit <= 0 {
		outputLimit = 10 // Default to a sensible limit if non-positive
	}
	return fmt.Sprintf("cat '%s' | tail -n %s | sed 's/[[:space:]]*$//' | sort | uniq -c | sort -nr | head -n %d", historyFilePath, historyScanCountStr, outputLimit), nil
}
