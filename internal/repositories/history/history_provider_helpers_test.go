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
	"github.com/AntonioJCosta/nicksh/internal/core/testutil"
)

// setupEnvVar sets an environment variable for the duration of the test.
func setupEnvVar(t *testing.T, key, value string) {
	t.Helper()
	originalValue, wasSet := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("Failed to set env var %s to %s: %v", key, value, err)
	}
	t.Cleanup(func() {
		if !wasSet {
			if err := os.Unsetenv(key); err != nil {
				t.Logf("Cleanup warning: failed to unset env var %s: %v", key, err)
			}
		} else {
			if err := os.Setenv(key, originalValue); err != nil {
				t.Logf("Cleanup warning: failed to restore env var %s to %s: %v", key, originalValue, err)
			}
		}
	})
}

// unsetEnvVar unsets an environment variable for the duration of the test.
func unsetEnvVar(t *testing.T, key string) {
	t.Helper()
	originalValue, wasSet := os.LookupEnv(key)
	if wasSet {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Failed to unset env var %s: %v", key, err)
		}
		t.Cleanup(func() {
			if err := os.Setenv(key, originalValue); err != nil {
				t.Logf("Cleanup warning: failed to restore env var %s to %s: %v", key, originalValue, err)
			}
		})
	}
}

// manageTestFile creates a file at the given path for the test and ensures it's cleaned up.
// If content is empty, an empty file is created.
func manageTestFile(t *testing.T, path string, content []byte) {
	t.Helper()
	// Ensure directory exists if path includes subdirectories in TempDir
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("Failed to create test file %s: %v", path, err)
	}
	t.Cleanup(func() {
		if err := os.Remove(path); err != nil {
			// Don't fail test on cleanup error, just log
			t.Logf("Warning: failed to remove test file %s: %v", path, err)
		}
	})
}

func TestFindUserHistoryFile(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}
	homeDir := currentUser.HomeDir
	tempDir := t.TempDir()

	// Define paths for mock files
	absHistFilePath := filepath.Join(tempDir, ".my_hist_absolute")
	relHistFilePathInHome := ".my_hist_relative"
	fullRelHistFilePathInHome := filepath.Join(homeDir, relHistFilePathInHome)

	defaultZshHistoryPath := filepath.Join(homeDir, ".zsh_history")
	defaultBashHistoryPath := filepath.Join(homeDir, ".bash_history")

	tests := []struct {
		name                  string
		providerSetup         func() *HistoryProvider // Allows for different provider states
		scanLimit             int
		outputLimit           int
		setupFunc             func(t *testing.T) // For setting env vars and creating files
		wantPath              string
		wantErr               bool
		wantErrorContain      string
		mockExecuteFunc       func(shellName, pipeline string) (string, string, error)
		wantFreqs             []history.CommandFrequency
		wantErrorContains     string
		checkPipelineContains []string // Optional: check parts of the generated pipeline
	}{
		{
			name: "HISTFILE set to existing absolute path",
			setupFunc: func(t *testing.T) {
				manageTestFile(t, absHistFilePath, []byte("cmd1"))
				setupEnvVar(t, "HISTFILE", absHistFilePath)
			},
			wantPath: absHistFilePath,
		},
		{
			name: "HISTFILE set to existing relative path (resolved from home)",
			setupFunc: func(t *testing.T) {
				manageTestFile(t, fullRelHistFilePathInHome, []byte("cmd1"))
				setupEnvVar(t, "HISTFILE", relHistFilePathInHome) // Relative path
			},
			wantPath: fullRelHistFilePathInHome,
		},
		{
			name: "HISTFILE set but file does not exist, .zsh_history exists",
			setupFunc: func(t *testing.T) {
				setupEnvVar(t, "HISTFILE", filepath.Join(tempDir, ".non_existent_histfile"))
				manageTestFile(t, defaultZshHistoryPath, []byte("zsh_cmd"))
			},
			wantPath: defaultZshHistoryPath,
		},
		{
			name: "HISTFILE not set, .zsh_history exists",
			setupFunc: func(t *testing.T) {
				unsetEnvVar(t, "HISTFILE")
				manageTestFile(t, defaultZshHistoryPath, []byte("zsh_cmd"))
			},
			wantPath: defaultZshHistoryPath,
		},
		{
			name: "HISTFILE not set, .zsh_history does not exist, .bash_history exists",
			setupFunc: func(t *testing.T) {
				unsetEnvVar(t, "HISTFILE")
				// Ensure .zsh_history does not exist by not creating it
				manageTestFile(t, defaultBashHistoryPath, []byte("bash_cmd"))
			},
			wantPath: defaultBashHistoryPath,
		},
		{
			name: "HISTFILE set to relative path not in home, file exists (should be resolved from home, thus not found, fallback)",
			setupFunc: func(t *testing.T) {
				// HISTFILE relative paths are resolved from home. If we set it to "temp/.hist"
				// it will look for "$HOME/temp/.hist".
				// We create a file at "$TEMPDIR/temp/.hist" which should NOT be found by HISTFILE logic.
				// Then it should fall back to default .zsh_history.
				tempSubDirHist := filepath.Join(tempDir, "temp_hist_file")
				manageTestFile(t, tempSubDirHist, []byte("cmd")) // This file won't be found by HISTFILE logic
				setupEnvVar(t, "HISTFILE", "some_other_relative_path_that_does_not_exist_in_home")
				manageTestFile(t, defaultZshHistoryPath, []byte("zsh_cmd")) // Fallback
			},
			wantPath: defaultZshHistoryPath,
		},
		{
			name: "No HISTFILE, no default files found",
			setupFunc: func(t *testing.T) {
				unsetEnvVar(t, "HISTFILE")
				// Ensure no default files exist by not creating them
			},
			wantErr:          true,
			wantErrorContain: "could not automatically find a common shell history file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			gotPath, err := findUserHistoryFile()

			if (err != nil) != tt.wantErr {
				t.Errorf("findUserHistoryFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrorContain) {
					t.Errorf("findUserHistoryFile() error %q does not contain %q", err.Error(), tt.wantErrorContain)
				}
				return
			}
			if gotPath != tt.wantPath {
				t.Errorf("findUserHistoryFile() = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestParsePipelineOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    []history.CommandFrequency
		wantErr bool // parsePipelineOutput currently doesn't return errors
	}{
		{
			name: "valid input",
			output: `
      5 git status
      3 ls -l
      1 echo hello
`,
			want: []history.CommandFrequency{
				{Command: "git status", Count: 5},
				{Command: "ls -l", Count: 3},
				{Command: "echo hello", Count: 1},
			},
		},
		{
			name:   "empty input",
			output: "",
			want:   []history.CommandFrequency{},
		},
		{
			name:   "input with only whitespace lines",
			output: "   \n   \n",
			want:   []history.CommandFrequency{},
		},
		{
			name: "input with invalid count, valid lines are still processed",
			output: `
      five git status
      3 ls -l
`,
			want: []history.CommandFrequency{
				{Command: "ls -l", Count: 3}, // "five git status" is skipped
			},
		},
		{
			name: "input with lines too short, valid lines are still processed",
			output: `
      5
      3 ls -l
`,
			want: []history.CommandFrequency{
				{Command: "ls -l", Count: 3}, // "5" is skipped
			},
		},
		{
			name:   "input with leading/trailing spaces in command part",
			output: `  2  my command  with  spaces  `,
			want: []history.CommandFrequency{
				{Command: "my command with spaces", Count: 2},
			},
		},
		{
			name:   "input with tabs instead of spaces",
			output: "\t7\tgit\tcommit\t-m\t'message'",
			want: []history.CommandFrequency{
				{Command: "git commit -m 'message'", Count: 7},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePipelineOutput(tt.output) // parsePipelineOutput itself doesn't return error
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePipelineOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePipelineOutput() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetermineScanCount(t *testing.T) {
	tests := []struct {
		name               string
		histSizeEnv        string // Value for HISTSIZE, or "__UNSET__"
		fcHistoryScanLimit int
		want               int
		// determineScanCount currently doesn't return errors, it defaults.
	}{
		{name: "HISTSIZE set and valid", histSizeEnv: "1000", fcHistoryScanLimit: 0, want: 1000},
		{name: "HISTSIZE set but invalid (not a number)", histSizeEnv: "abc", fcHistoryScanLimit: 0, want: 500}, // Default
		{name: "HISTSIZE set but invalid (zero)", histSizeEnv: "0", fcHistoryScanLimit: 0, want: 500},           // Default
		{name: "HISTSIZE set but invalid (negative)", histSizeEnv: "-100", fcHistoryScanLimit: 0, want: 500},    // Default
		{name: "HISTSIZE not set, fcHistoryScanLimit positive", histSizeEnv: "__UNSET__", fcHistoryScanLimit: 200, want: 200},
		{name: "HISTSIZE not set, fcHistoryScanLimit zero", histSizeEnv: "__UNSET__", fcHistoryScanLimit: 0, want: 500},       // Default
		{name: "HISTSIZE not set, fcHistoryScanLimit negative", histSizeEnv: "__UNSET__", fcHistoryScanLimit: -10, want: 500}, // Default
		{name: "fcHistoryScanLimit positive takes precedence over HISTSIZE", histSizeEnv: "1000", fcHistoryScanLimit: 200, want: 200},
		{name: "fcHistoryScanLimit zero, HISTSIZE valid", histSizeEnv: "1000", fcHistoryScanLimit: 0, want: 1000},
		{name: "fcHistoryScanLimit negative, HISTSIZE valid", histSizeEnv: "1000", fcHistoryScanLimit: -5, want: 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.histSizeEnv == "__UNSET__" {
				unsetEnvVar(t, "HISTSIZE")
			} else {
				setupEnvVar(t, "HISTSIZE", tt.histSizeEnv)
			}

			// determineScanCount doesn't return an error in the current implementation
			got, _ := determineScanCount(tt.fcHistoryScanLimit)
			if got != tt.want {
				t.Errorf("determineScanCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildShellPipeline(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "history_for_pipeline_")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close() // Close immediately, we just need the path for os.Stat

	nonExistentFilePath := filepath.Join(t.TempDir(), "non_existent_history_file")

	tests := []struct {
		name                string
		historyFilePath     string
		historyScanCountStr string
		outputLimit         int
		wantPipelineParts   []string // Substrings the pipeline should contain
		wantErr             bool
		wantErrorContain    string
	}{
		{
			name:                "generic pipeline construction",
			historyFilePath:     tmpFilePath,
			historyScanCountStr: "1000",
			outputLimit:         10,
			wantPipelineParts:   []string{fmt.Sprintf("cat '%s'", tmpFilePath), "tail -n 1000", "sed 's/[[:space:]]*$//'", "sort | uniq -c | sort -nr", "head -n 10"},
		},
		{
			name:                "different scan count and output limit",
			historyFilePath:     tmpFilePath,
			historyScanCountStr: "500",
			outputLimit:         5,
			wantPipelineParts:   []string{fmt.Sprintf("cat '%s'", tmpFilePath), "tail -n 500", "sed 's/[[:space:]]*$//'", "sort | uniq -c | sort -nr", "head -n 5"},
		},
		{
			name:                "output limit zero or negative (defaults to 10)",
			historyFilePath:     tmpFilePath,
			historyScanCountStr: "100",
			outputLimit:         0,
			wantPipelineParts:   []string{"head -n 10"},
		},
		{
			name:                "history file does not exist",
			historyFilePath:     nonExistentFilePath,
			historyScanCountStr: "100",
			outputLimit:         10,
			wantErr:             true,
			wantErrorContain:    fmt.Sprintf("history file does not exist: %s", toUserFriendlyPath(nonExistentFilePath)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure the temp file exists for non-error cases, or doesn't for error cases
			if !tt.wantErr && tt.historyFilePath == tmpFilePath {
				manageTestFile(t, tmpFilePath, []byte("content"))
			} else if tt.wantErr && tt.historyFilePath == nonExistentFilePath {
				os.Remove(nonExistentFilePath) // Ensure it doesn't exist
			}

			gotPipeline, err := buildShellPipeline(tt.historyFilePath, tt.historyScanCountStr, tt.outputLimit)

			if (err != nil) != tt.wantErr {
				t.Errorf("buildShellPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrorContain) {
					t.Errorf("buildShellPipeline() error %q does not contain %q", err.Error(), tt.wantErrorContain)
				}
				return
			}
			for _, part := range tt.wantPipelineParts {
				if !strings.Contains(gotPipeline, part) {
					t.Errorf("buildShellPipeline() = %q, does not contain %q", gotPipeline, part)
				}
			}
		})
	}
}

func TestHistoryProvider_getHistoryFrequencies(t *testing.T) {
	tmpHistoryFile, err := os.CreateTemp(t.TempDir(), "provider_history_")
	if err != nil {
		t.Fatalf("Failed to create temp history file: %v", err)
	}
	historyFilePath := tmpHistoryFile.Name()
	manageTestFile(t, historyFilePath, []byte("some command\nanother command\nsome command\n")) // Ensure file exists and has content
	// Note: manageTestFile handles cleanup

	tests := []struct {
		name                  string
		providerSetup         func() *HistoryProvider // Allows for different provider states
		scanLimit             int
		outputLimit           int
		mockExecuteFunc       func(shellName, pipeline string) (string, string, error)
		wantFreqs             []history.CommandFrequency
		wantErr               bool
		wantErrorContains     string
		checkPipelineContains []string // Optional: check parts of the generated pipeline
	}{
		{
			name: "successful pipeline execution",
			providerSetup: func() *HistoryProvider {
				return &HistoryProvider{Shell: "bash", HistoryFile: historyFilePath}
			},
			scanLimit:   100,
			outputLimit: 10,
			mockExecuteFunc: func(shellName, pipeline string) (string, string, error) {
				return "2 some command\n1 another command", "", nil
			},
			wantFreqs: []history.CommandFrequency{
				{Command: "some command", Count: 2},
				{Command: "another command", Count: 1},
			},
			checkPipelineContains: []string{"tail -n 100", "head -n 10"}, // scanLimit is 100 (from HISTSIZE or default if fcHistoryScanLimit is 0)
		},
		{
			name: "pipeline execution error",
			providerSetup: func() *HistoryProvider {
				return &HistoryProvider{Shell: "zsh", HistoryFile: historyFilePath}
			},
			scanLimit:   50,
			outputLimit: 5,
			mockExecuteFunc: func(shellName, pipeline string) (string, string, error) {
				return "", "permission denied", errors.New("exit status 1")
			},
			wantErr: true,
		},
		{
			name: "history file not set in provider",
			providerSetup: func() *HistoryProvider {
				return &HistoryProvider{Shell: "bash", HistoryFile: ""} // HistoryFile is empty
			},
			scanLimit:   100,
			outputLimit: 10,
			mockExecuteFunc: func(shellName, pipeline string) (string, string, error) {
				t.Fatal("cmdExecutor.Execute should not be called if HistoryFile is not set")
				return "", "", nil
			},
			wantErr:           true,
			wantErrorContains: "history file path is not set",
		},
		{
			name: "buildShellPipeline fails (e.g., history file becomes non-existent after provider init)",
			providerSetup: func() *HistoryProvider {
				// Simulate file existing at init, but removed before getHistoryFrequencies call
				// For this test, we'll set a path that buildShellPipeline will fail on.
				return &HistoryProvider{Shell: "bash", HistoryFile: filepath.Join(t.TempDir(), "file_will_be_gone")}
			},
			scanLimit:   100,
			outputLimit: 10,
			mockExecuteFunc: func(shellName, pipeline string) (string, string, error) {
				t.Fatal("cmdExecutor.Execute should not be called if buildShellPipeline fails")
				return "", "", nil
			},
			wantErr:           true,
			wantErrorContains: "building shell pipeline: history file does not exist",
		},
		{
			name: "parsePipelineOutput yields empty due to malformed executor output",
			providerSetup: func() *HistoryProvider {
				return &HistoryProvider{Shell: "bash", HistoryFile: historyFilePath}
			},
			scanLimit:   100,
			outputLimit: 10,
			mockExecuteFunc: func(shellName, pipeline string) (string, string, error) {
				return "this is not valid output format", "", nil // Malformed output
			},
			wantFreqs: []history.CommandFrequency{}, // Expect empty slice, not an error
		},
		// Note on the potential bug in source:
		// If cmdExecutor.Execute returns an error, and stdout is empty,
		// the current source code returns: errors.New("history file path is not set in HistoryProvider")
		// This shadows the actual execution error. The tests here assume this bug might be fixed
		// to return the more detailed error (like in "pipeline execution error" case).
		// If the bug persists, the wantErrorContains for such a specific case would need to be
		// "history file path is not set in HistoryProvider".
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tt.providerSetup()
			provider.cmdExecutor = &testutil.MockCommandExecutor{
				ExecuteFunc: func(shellName, pipeline string) (string, string, error) {
					if tt.checkPipelineContains != nil {
						for _, part := range tt.checkPipelineContains {
							if !strings.Contains(pipeline, part) {
								t.Errorf("Pipeline %q was expected to contain %q", pipeline, part)
							}
						}
					}
					if tt.mockExecuteFunc != nil {
						return tt.mockExecuteFunc(shellName, pipeline)
					}
					return "", "", errors.New("mockExecuteFunc not provided")
				},
			}

			// Special handling for the "buildShellPipeline fails" case
			// Ensure the file does not exist if that's what we're testing for buildShellPipeline
			if tt.name == "buildShellPipeline fails (e.g., history file becomes non-existent after provider init)" {
				os.Remove(provider.HistoryFile)
			}

			freqs, err := provider.getHistoryFrequencies(tt.scanLimit, tt.outputLimit)

			if (err != nil) != tt.wantErr {
				t.Errorf("getHistoryFrequencies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrorContains) {
					t.Errorf("getHistoryFrequencies() error %q does not contain %q", err.Error(), tt.wantErrorContains)
				}
				return
			}
			if !reflect.DeepEqual(freqs, tt.wantFreqs) {
				t.Errorf("getHistoryFrequencies() freqs = %#v, want %#v", freqs, tt.wantFreqs)
			}
		})
	}
}
