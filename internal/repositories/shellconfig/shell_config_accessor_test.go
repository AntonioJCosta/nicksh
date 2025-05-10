package shellconfig

import (
	"io"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
)

// Helper to manage environment variables for tests
func setupEnvVar(t *testing.T, key, value string) {
	t.Helper()
	originalValue, isset := os.LookupEnv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if isset {
			os.Setenv(key, originalValue)
		} else {
			os.Unsetenv(key)
		}
	})
}

func unsetEnvVar(t *testing.T, key string) {
	t.Helper()
	originalValue, isset := os.LookupEnv(key)
	os.Unsetenv(key)
	t.Cleanup(func() {
		if isset {
			os.Setenv(key, originalValue)
		}
	})
}

// manageTestFile is already defined in shell_config_accessor_helpers_test.go
// If it's not in the same package for testing, it needs to be redefined or imported.
// For this example, we assume it's accessible or redefined here if needed.
// func manageTestFile(t *testing.T, path string, content []byte) { ... }

func TestNewShellConfigAccessor(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user for test setup: %v", err)
	}
	expectedHomeDir := currentUser.HomeDir

	tests := []struct {
		name                     string
		setupFunc                func()
		wantErr                  bool
		wantErrorContains        string
		expectedShell            string
		expectedGenAliasesPathFn func(home string) string
	}{
		{
			name: "SHELL variable set",
			setupFunc: func() {
				setupEnvVar(t, "SHELL", "/bin/zsh")
			},
			wantErr:       false,
			expectedShell: "zsh",
			expectedGenAliasesPathFn: func(home string) string {
				return filepath.Join(home, ".nicksh", "generated_aliases")
			},
		},
		{
			name: "SHELL variable set with path",
			setupFunc: func() {
				setupEnvVar(t, "SHELL", "/usr/local/bin/bash")
			},
			wantErr:       false,
			expectedShell: "bash",
			expectedGenAliasesPathFn: func(home string) string {
				return filepath.Join(home, ".nicksh", "generated_aliases")
			},
		},
		{
			name: "SHELL variable not set",
			setupFunc: func() {
				unsetEnvVar(t, "SHELL")
			},
			wantErr:           true,
			wantErrorContains: "SHELL environment variable not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			accessor, err := NewShellConfigAccessor()

			if (err != nil) != tt.wantErr {
				t.Errorf("NewShellConfigAccessor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrorContains) {
					t.Errorf("NewShellConfigAccessor() error = %q, want error containing %q", err.Error(), tt.wantErrorContains)
				}
				return
			}

			if accessor == nil {
				t.Fatal("NewShellConfigAccessor() returned nil accessor on success")
			}

			sca, ok := accessor.(*ShellConfigAccessor)
			if !ok {
				t.Fatal("NewShellConfigAccessor() did not return a *ShellConfigAccessor")
			}

			if sca.shell != tt.expectedShell {
				t.Errorf("NewShellConfigAccessor() shell = %q, want %q", sca.shell, tt.expectedShell)
			}

			expectedPath := tt.expectedGenAliasesPathFn(expectedHomeDir)
			if sca.generatedAliasesFilePath != expectedPath {
				t.Errorf("NewShellConfigAccessor() generatedAliasesFilePath = %q, want %q", sca.generatedAliasesFilePath, expectedPath)
			}
		})
	}
}

func TestShellConfigAccessor_GetExistingAliases(t *testing.T) {
	baseTempDir := t.TempDir() // Base for all test-specific "home" dirs

	tests := []struct {
		name           string
		setupFiles     func(aliasesDir string) // Function to set up files in the aliasesDir
		expectedOutput map[string]string
		wantErr        bool
		wantErrMsg     string
		expectedStderr string
	}{
		{
			name: "alias directory does not exist",
			setupFiles: func(aliasesDir string) {
				// Do nothing, directory won't be created
			},
			expectedOutput: map[string]string{},
			wantErr:        false,
		},
		{
			name: "alias directory exists but is empty",
			setupFiles: func(aliasesDir string) {
				if err := os.MkdirAll(aliasesDir, 0755); err != nil {
					t.Fatalf("Failed to create aliasesDir: %v", err)
				}
			},
			expectedOutput: map[string]string{},
			wantErr:        false,
		},
		{
			name: "alias directory with one file",
			setupFiles: func(aliasesDir string) {
				if err := os.MkdirAll(aliasesDir, 0755); err != nil {
					t.Fatalf("Failed to create aliasesDir: %v", err)
				}
				manageTestFile(t, filepath.Join(aliasesDir, "file1.aliases"), []byte("alias g=git\nalias ll='ls -l'"))
			},
			expectedOutput: map[string]string{"g": "git", "ll": "ls -l"},
			wantErr:        false,
		},
		{
			name: "alias directory with multiple files, no conflicts",
			setupFiles: func(aliasesDir string) {
				if err := os.MkdirAll(aliasesDir, 0755); err != nil {
					t.Fatalf("Failed to create aliasesDir: %v", err)
				}
				manageTestFile(t, filepath.Join(aliasesDir, "file1.aliases"), []byte("alias g=git"))
				manageTestFile(t, filepath.Join(aliasesDir, "file2.aliases"), []byte("alias ll='ls -l'"))
			},
			expectedOutput: map[string]string{"g": "git", "ll": "ls -l"},
			wantErr:        false,
		},
		{
			name: "alias directory with multiple files, with conflicts (last wins)",
			setupFiles: func(aliasesDir string) {
				if err := os.MkdirAll(aliasesDir, 0755); err != nil {
					t.Fatalf("Failed to create aliasesDir: %v", err)
				}
				// Order of ReadDir is not guaranteed, but we can check for the warning
				// and that one of them wins. Let's assume file1 is read then file2.
				manageTestFile(t, filepath.Join(aliasesDir, "file1.aliases"), []byte("alias c=cmd1"))
				manageTestFile(t, filepath.Join(aliasesDir, "file2.aliases"), []byte("alias c=cmd2\nalias k=kubectl"))
			},
			expectedOutput: map[string]string{"c": "cmd2", "k": "kubectl"}, // cmd2 from file2 should win
			wantErr:        false,
			expectedStderr: "Warning: Alias 'c' found in multiple files.", // Check if warning is logged
		},
		{
			name: "alias directory with a file that causes getAliasesFromFile to error",
			setupFiles: func(aliasesDir string) {
				if err := os.MkdirAll(aliasesDir, 0755); err != nil {
					t.Fatalf("Failed to create aliasesDir: %v", err)
				}
				manageTestFile(t, filepath.Join(aliasesDir, "good.aliases"), []byte("alias g=git"))
				// Create a file that might cause scanning error if getAliasesFromFile was more complex
				// For now, getAliasesFromFile handles os.Open errors.
				// Let's simulate a problematic file by making it a directory (os.Open will fail)
				if err := os.Mkdir(filepath.Join(aliasesDir, "badfile.aliases"), 0755); err != nil {
					t.Fatalf("failed to create badfile.aliases dir: %v", err)
				}
			},
			expectedOutput: map[string]string{"g": "git"}, // Should still get aliases from good.aliases
			wantErr:        false,
		},
		{
			name: "os.ReadDir fails for the alias directory",
			setupFiles: func(aliasesDir string) {
				// Create the aliasesDir path as a file, so os.ReadDir fails
				manageTestFile(t, aliasesDir, []byte("this is not a directory"))
			},
			expectedOutput: nil,
			wantErr:        true,
			wantErrMsg:     "failed to read alias directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a unique "home" for this test run to place .nicksh
			testHomeDir := filepath.Join(baseTempDir, strings.ReplaceAll(tt.name, " ", "_"))
			if err := os.MkdirAll(testHomeDir, 0755); err != nil {
				t.Fatalf("Failed to create testHomeDir: %v", err)
			}

			aliasesDir := filepath.Join(testHomeDir, generatedAliasesDir)
			tt.setupFiles(aliasesDir)

			sca := &ShellConfigAccessor{
				shell:                    "testshell",
				generatedAliasesFilePath: filepath.Join(aliasesDir, generatedAliasesFilename), // Path used by AddAlias, GetExistingAliases uses its dir
			}

			// Capture stderr
			oldStderr := os.Stderr
			rErr, wErr, _ := os.Pipe()
			os.Stderr = wErr
			defer func() {
				os.Stderr = oldStderr
				wErr.Close()
				rErr.Close()
			}()

			aliases, err := sca.GetExistingAliases()

			wErr.Close() // Close writer to allow reader to get EOF
			stderrBytes, _ := io.ReadAll(rErr)
			stderrOutput := string(stderrBytes)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetExistingAliases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("GetExistingAliases() error = %q, want to contain %q", err.Error(), tt.wantErrMsg)
				}
			}
			if !reflect.DeepEqual(aliases, tt.expectedOutput) {
				t.Errorf("GetExistingAliases() aliases = %v, want %v", aliases, tt.expectedOutput)
			}
			if tt.expectedStderr != "" && !strings.Contains(stderrOutput, tt.expectedStderr) {
				t.Errorf("GetExistingAliases() stderr = %q, want to contain %q", stderrOutput, tt.expectedStderr)
			}
		})
	}
}

func TestShellConfigAccessor_AddAlias(t *testing.T) {
	baseTempDir := t.TempDir()

	tests := []struct {
		name                string
		initialFileContent  *string // Pointer to distinguish between no file and empty file
		aliasToAdd          alias.Alias
		expectedAdded       bool
		expectedFileContent string
		wantErr             bool
		wantErrMsg          string
		expectedStdout      string
	}{
		{
			name:                "add to non-existent file",
			initialFileContent:  nil, // File doesn't exist
			aliasToAdd:          alias.Alias{Name: "g", Command: "git"},
			expectedAdded:       true,
			expectedFileContent: "alias g='git'\n",
			wantErr:             false,
			expectedStdout:      "Alias 'g' added to",
		},
		{
			name:                "add to existing empty file",
			initialFileContent:  stringp(""),
			aliasToAdd:          alias.Alias{Name: "ll", Command: "ls -l"},
			expectedAdded:       true,
			expectedFileContent: "alias ll='ls -l'\n",
			wantErr:             false,
			expectedStdout:      "Alias 'll' added to",
		},
		{
			name:                "add to existing file with content",
			initialFileContent:  stringp("alias k=kubectl\n"),
			aliasToAdd:          alias.Alias{Name: "gp", Command: "git push"},
			expectedAdded:       true,
			expectedFileContent: "alias k=kubectl\nalias gp='git push'\n",
			wantErr:             false,
			expectedStdout:      "Alias 'gp' added to",
		},
		{
			name:                "add alias that already exists",
			initialFileContent:  stringp("alias g='git'\n"),
			aliasToAdd:          alias.Alias{Name: "g", Command: "git status"}, // Same name
			expectedAdded:       false,
			expectedFileContent: "alias g='git'\n", // File should not change
			wantErr:             false,
			expectedStdout:      "Alias 'g' already exists",
		},
		// Error cases for os.MkdirAll, os.OpenFile, file.WriteString are harder to test
		// without more complex mocking of os-level functions or specific file system states.
		// For example, to test MkdirAll failure, the parent path would need to be a file.
		// To test OpenFile failure, the file could be a directory.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testHomeDir := filepath.Join(baseTempDir, strings.ReplaceAll(tt.name, " ", "_"))
			if err := os.MkdirAll(testHomeDir, 0755); err != nil {
				t.Fatalf("Failed to create testHomeDir: %v", err)
			}

			aliasesDir := filepath.Join(testHomeDir, generatedAliasesDir)
			generatedFile := filepath.Join(aliasesDir, generatedAliasesFilename)

			sca := &ShellConfigAccessor{
				shell:                    "testshell",
				generatedAliasesFilePath: generatedFile,
			}

			if tt.initialFileContent != nil { // Setup initial file if specified
				if err := os.MkdirAll(aliasesDir, 0755); err != nil {
					t.Fatalf("Failed to create aliasesDir for initial content: %v", err)
				}
				manageTestFile(t, generatedFile, []byte(*tt.initialFileContent))
			}

			// Capture stdout
			oldStdout := os.Stdout
			rOut, wOut, _ := os.Pipe()
			os.Stdout = wOut
			defer func() {
				os.Stdout = oldStdout
				wOut.Close()
				rOut.Close()
			}()

			added, err := sca.AddAlias(tt.aliasToAdd)

			wOut.Close() // Close writer to allow reader to get EOF
			stdoutBytes, _ := io.ReadAll(rOut)
			stdoutOutput := string(stdoutBytes)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddAlias() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("AddAlias() error = %q, want to contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if added != tt.expectedAdded {
				t.Errorf("AddAlias() added = %v, want %v", added, tt.expectedAdded)
			}

			if tt.expectedAdded || (!tt.expectedAdded && tt.initialFileContent != nil && *tt.initialFileContent == tt.expectedFileContent) { // Check content if added or if skipped and content should remain same
				fileContentBytes, readErr := os.ReadFile(generatedFile)
				if readErr != nil && tt.expectedAdded { // If we expected to add, file should exist
					t.Fatalf("Failed to read generated aliases file %s: %v", generatedFile, readErr)
				}
				if readErr == nil { // Only compare if file was readable (or expected to exist)
					fileContent := string(fileContentBytes)
					if fileContent != tt.expectedFileContent {
						t.Errorf("AddAlias() file content = %q, want %q", fileContent, tt.expectedFileContent)
					}
				} else if tt.expectedFileContent != "" { // If we expected content but file not readable
					t.Errorf("AddAlias() expected file content %q, but file not readable: %v", tt.expectedFileContent, readErr)
				}
			}

			if tt.expectedStdout != "" && !strings.Contains(stdoutOutput, tt.expectedStdout) {
				t.Errorf("AddAlias() stdout = %q, want to contain %q", stdoutOutput, tt.expectedStdout)
			}

			// Cleanup the specific generated file if it was created and no initial content was set
			// This helps if the test failed before t.Cleanup on TempDir runs.
			if tt.initialFileContent == nil {
				os.Remove(generatedFile)
				os.Remove(aliasesDir) // Try to remove dir too
			}
		})
	}
}

// stringp returns a pointer to a string.
func stringp(s string) *string {
	return &s
}
