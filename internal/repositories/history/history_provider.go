package history

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/history"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
)

/*
HistoryProvider provides access to shell command history stored in files.
It implements the ports.HistoryProvider interface.
*/
type HistoryProvider struct {
	Shell            string
	HistoryFile      string // Stores the absolute path
	cmdExecutor      ports.CommandExecutor
	sourceIdentifier string // Stores the user-friendly source identifier
}

func (hp *HistoryProvider) GetSourceIdentifier() string {
	if hp.sourceIdentifier != "" {
		return hp.sourceIdentifier
	}
	// Fallback logic if sourceIdentifier was not pre-computed (should ideally not be needed with current NewHistoryProvider)
	if hp.HistoryFile != "" {
		return fmt.Sprintf("File: %s", toUserFriendlyPath(hp.HistoryFile))
	}
	return fmt.Sprintf("Shell: %s (history file path unknown)", hp.Shell)
}

// NewHistoryProvider creates a new FileBasedHistoryProvider.
func NewHistoryProvider(cmdExecutor ports.CommandExecutor, fileFinder ports.HistoryFileFinder) (ports.HistoryProvider, error) {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return nil, fmt.Errorf("SHELL environment variable not set")
	}

	shellName := strings.ToLower(filepath.Base(shellPath))
	histFilePath, err := fileFinder.Find()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not automatically find a history file: %v. History-based suggestions might be unavailable.\n", err)
		return &HistoryProvider{
			Shell:            shellName,
			cmdExecutor:      cmdExecutor,
			sourceIdentifier: fmt.Sprintf("Shell: %s (history file not found or configured)", shellName),
		}, nil
	}

	userFriendlyHistPath := toUserFriendlyPath(histFilePath)

	return &HistoryProvider{
		HistoryFile:      histFilePath, // Store the actual absolute path for internal use
		Shell:            shellName,
		cmdExecutor:      cmdExecutor,
		sourceIdentifier: fmt.Sprintf("File: %s", userFriendlyHistPath), // Store user-friendly path for display
	}, nil
}

// GetCommandFrequencies implements the ports.HistoryProvider interface.
func (hp *HistoryProvider) GetCommandFrequencies(scanLimit int, outputLimit int) ([]history.CommandFrequency, error) {
	if hp.HistoryFile == "" {
		return nil, fmt.Errorf("history file not found or configured for shell %s. Cannot fetch command frequencies", hp.Shell)
	}

	return hp.getHistoryFrequencies(scanLimit, outputLimit)
}

func (hp *HistoryProvider) GetHistoryFilePath() string {
	return hp.HistoryFile
}
