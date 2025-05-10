package history

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/history"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
	"github.com/AntonioJCosta/nicksh/internal/core/testutil"
)

// Re-define or import helpers if they are not in the same package
// For simplicity, assuming these helpers from history_provider_helpers_test.go are accessible
// or can be copied/adapted if in a different test package.
// If they are in the same package `history`, they can be used directly.
func TestNewHistoryProvider(t *testing.T) {
	currentUser, _ := user.Current()
	homeDir := currentUser.HomeDir
	tempDir := t.TempDir() // For generating realistic temp paths if needed by mocks
	mockCmdExecutor := &testutil.MockCommandExecutor{}

	tests := []struct {
		name                  string
		setupShellEnv         func()
		mockFileFinder        ports.HistoryFileFinder // Use the mock interface
		wantProviderNonNil    bool
		wantErr               bool
		wantErrorContains     string
		wantHistoryFile       string // Expected absolute path
		wantSourceIdentifier  string
		checkSourceIdentifier bool
	}{
		{
			name: "SHELL not set",
			setupShellEnv: func() {
				unsetEnvVar(t, "SHELL")
			},
			mockFileFinder:     &testutil.MockHistoryFileFinder{FindFunc: func() (string, error) { return "", nil }}, // Should not be called if SHELL fails first
			wantProviderNonNil: false,
			wantErr:            true,
			wantErrorContains:  "SHELL environment variable not set",
		},
		{
			name: "SHELL set, history file found by mock",
			setupShellEnv: func() {
				setupEnvVar(t, "SHELL", "/bin/zsh")
			},
			mockFileFinder: &testutil.MockHistoryFileFinder{
				FindFunc: func() (string, error) {
					return filepath.Join(homeDir, ".zsh_history"), nil
				},
			},
			wantProviderNonNil:    true,
			wantErr:               false,
			wantHistoryFile:       filepath.Join(homeDir, ".zsh_history"),
			wantSourceIdentifier:  fmt.Sprintf("File: %s", toUserFriendlyPath(filepath.Join(homeDir, ".zsh_history"))),
			checkSourceIdentifier: true,
		},
		{
			name: "SHELL set, history file not found by mock (error returned)",
			setupShellEnv: func() {
				setupEnvVar(t, "SHELL", "/bin/bash")
			},
			mockFileFinder: &testutil.MockHistoryFileFinder{
				FindFunc: func() (string, error) {
					return "", errors.New("mock: no history file found")
				},
			},
			wantProviderNonNil:    true,
			wantErr:               false,
			wantHistoryFile:       "",
			wantSourceIdentifier:  "Shell: bash (history file not found or configured)",
			checkSourceIdentifier: true,
		},
		{
			name: "SHELL set, HISTFILE points to a temp file (using DefaultFileFinder for this specific case to test integration)",
			setupShellEnv: func() {
				setupEnvVar(t, "SHELL", "/usr/bin/fish")
				histFile := filepath.Join(tempDir, ".my_custom_hist")
				manageTestFile(t, histFile, []byte("cmd1")) // manageTestFile needs to be accessible
				setupEnvVar(t, "HISTFILE", histFile)        // Set HISTFILE for the real findUserHistoryFile
			},
			// For this one case, we can use the real DefaultHistoryFileFinder
			// to ensure the HISTFILE logic within findUserHistoryFile is covered.
			// This assumes findUserHistoryFile is still a global function called by DefaultHistoryFileFinder.
			// If findUserHistoryFile itself was refactored into DefaultHistoryFileFinder, this is fine.
			mockFileFinder:        NewDefaultHistoryFileFinder(),
			wantProviderNonNil:    true,
			wantErr:               false,
			wantHistoryFile:       filepath.Join(tempDir, ".my_custom_hist"),
			wantSourceIdentifier:  fmt.Sprintf("File: %s", toUserFriendlyPath(filepath.Join(tempDir, ".my_custom_hist"))),
			checkSourceIdentifier: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupShellEnv()

			// Redirect stderr to capture warning
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			provider, err := NewHistoryProvider(mockCmdExecutor, tt.mockFileFinder)

			w.Close()
			// capturedStderrBytes, _ := io.ReadAll(r) // If you need to assert warnings
			os.Stderr = oldStderr
			r.Close()

			if (err != nil) != tt.wantErr {
				t.Errorf("NewHistoryProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrorContains) {
					t.Errorf("NewHistoryProvider() error = %q, want error containing %q", err.Error(), tt.wantErrorContains)
				}
				return
			}
			if tt.wantProviderNonNil && provider == nil {
				t.Errorf("NewHistoryProvider() expected a non-nil provider, got nil")
				return
			}
			if !tt.wantProviderNonNil && provider != nil {
				t.Errorf("NewHistoryProvider() expected a nil provider, got non-nil")
				return
			}

			if provider != nil {
				hp, ok := provider.(*HistoryProvider)
				if !ok {
					t.Fatalf("NewHistoryProvider() did not return a *HistoryProvider")
				}
				if hp.HistoryFile != tt.wantHistoryFile {
					t.Errorf("NewHistoryProvider() HistoryFile = %q, want %q", hp.HistoryFile, tt.wantHistoryFile)
				}
				if tt.checkSourceIdentifier && hp.GetSourceIdentifier() != tt.wantSourceIdentifier {
					t.Errorf("NewHistoryProvider() SourceIdentifier = %q, want %q", hp.GetSourceIdentifier(), tt.wantSourceIdentifier)
				}
				if hp.cmdExecutor == nil {
					t.Error("NewHistoryProvider() cmdExecutor is nil")
				}
			}
		})
	}
}

func TestHistoryProvider_GetSourceIdentifier(t *testing.T) {
	currentUser, _ := user.Current()
	homeDir := currentUser.HomeDir
	absPath := filepath.Join(homeDir, ".zsh_history")
	userFriendlyAbsPath := toUserFriendlyPath(absPath)

	tests := []struct {
		name           string
		provider       *HistoryProvider
		wantIdentifier string
	}{
		{
			name: "Identifier set during creation (file found)",
			provider: &HistoryProvider{
				HistoryFile:      absPath,
				Shell:            "zsh",
				sourceIdentifier: fmt.Sprintf("File: %s", userFriendlyAbsPath),
			},
			wantIdentifier: fmt.Sprintf("File: %s", userFriendlyAbsPath),
		},
		{
			name: "Identifier set during creation (file not found)",
			provider: &HistoryProvider{
				HistoryFile:      "",
				Shell:            "bash",
				sourceIdentifier: "Shell: bash (history file not found or configured)",
			},
			wantIdentifier: "Shell: bash (history file not found or configured)",
		},
		{
			name: "Fallback: sourceIdentifier empty, HistoryFile set",
			provider: &HistoryProvider{ // Manually create to test fallback
				HistoryFile:      absPath,
				Shell:            "zsh",
				sourceIdentifier: "", // Force empty
			},
			wantIdentifier: fmt.Sprintf("File: %s", userFriendlyAbsPath),
		},
		{
			name: "Fallback: sourceIdentifier empty, HistoryFile empty",
			provider: &HistoryProvider{ // Manually create to test fallback
				HistoryFile:      "",
				Shell:            "fish",
				sourceIdentifier: "", // Force empty
			},
			wantIdentifier: "Shell: fish (history file path unknown)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.provider.GetSourceIdentifier(); got != tt.wantIdentifier {
				t.Errorf("GetSourceIdentifier() = %q, want %q", got, tt.wantIdentifier)
			}
		})
	}
}

func TestHistoryProvider_GetCommandFrequencies(t *testing.T) {
	tempDir := t.TempDir()
	historyFilePath := filepath.Join(tempDir, ".test_history")
	manageTestFile(t, historyFilePath, []byte("cmd1\ncmd2\ncmd1"))

	mockExecutor := &testutil.MockCommandExecutor{}
	providerWithFile := &HistoryProvider{
		Shell:            "bash",
		HistoryFile:      historyFilePath,
		cmdExecutor:      mockExecutor,
		sourceIdentifier: fmt.Sprintf("File: %s", toUserFriendlyPath(historyFilePath)),
	}
	providerWithoutFile := &HistoryProvider{
		Shell:            "zsh",
		HistoryFile:      "", // No history file
		cmdExecutor:      mockExecutor,
		sourceIdentifier: "Shell: zsh (history file not found or configured)",
	}

	tests := []struct {
		name              string
		provider          ports.HistoryProvider
		setupMockExecutor func()
		scanLimit         int
		outputLimit       int
		wantFreqs         []history.CommandFrequency
		wantErr           bool
		wantErrorContains string
	}{
		{
			name:              "HistoryFile not set on provider",
			provider:          providerWithoutFile,
			setupMockExecutor: func() {},
			scanLimit:         100,
			outputLimit:       10,
			wantErr:           true,
			wantErrorContains: "history file not found or configured",
		},
		{
			name:     "Successful fetch",
			provider: providerWithFile,
			setupMockExecutor: func() {
				mockExecutor.ExecuteFunc = func(shellName, pipeline string) (string, string, error) {
					return "2 cmd1\n1 cmd2", "", nil
				}
			},
			scanLimit:   100,
			outputLimit: 10,
			wantFreqs: []history.CommandFrequency{
				{Command: "cmd1", Count: 2},
				{Command: "cmd2", Count: 1},
			},
			wantErr: false,
		},
		{
			name:     "Error from command executor",
			provider: providerWithFile,
			setupMockExecutor: func() {
				mockExecutor.ExecuteFunc = func(shellName, pipeline string) (string, string, error) {
					return "", "exec error", errors.New("pipeline failed")
				}
			},
			scanLimit:   100,
			outputLimit: 10,
			wantErr:     true,
		},
		{
			name:     "Error from command executor with empty stdout (specific error shadowing case)",
			provider: providerWithFile,
			setupMockExecutor: func() {
				mockExecutor.ExecuteFunc = func(shellName, pipeline string) (string, string, error) {
					return "", "", errors.New("pipeline failed with empty stdout") // stdout is empty
				}
			},
			scanLimit:         100,
			outputLimit:       10,
			wantErr:           true,
			wantErrorContains: "history file path is not set in HistoryProvider", // This reflects the current behavior noted in history_provider_helpers_test.go
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMockExecutor()
			freqs, err := tt.provider.GetCommandFrequencies(tt.scanLimit, tt.outputLimit)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommandFrequencies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrorContains) {
					t.Errorf("GetCommandFrequencies() error = %q, want error containing %q", err.Error(), tt.wantErrorContains)
				}
				return
			}
			if !reflect.DeepEqual(freqs, tt.wantFreqs) {
				t.Errorf("GetCommandFrequencies() freqs = %v, want %v", freqs, tt.wantFreqs)
			}
		})
	}
}

func TestHistoryProvider_GetHistoryFilePath(t *testing.T) {
	absPath := "/home/user/.zsh_history"
	tests := []struct {
		name     string
		provider *HistoryProvider
		wantPath string
	}{
		{
			name: "Path is set",
			provider: &HistoryProvider{
				HistoryFile: absPath,
			},
			wantPath: absPath,
		},
		{
			name: "Path is not set",
			provider: &HistoryProvider{
				HistoryFile: "",
			},
			wantPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.provider.GetHistoryFilePath(); got != tt.wantPath {
				t.Errorf("GetHistoryFilePath() = %q, want %q", got, tt.wantPath)
			}
		})
	}
}
