package predefinedaliases

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
)

// Helper function to create a temporary YAML file for testing.
func createTempYAMLFile(t *testing.T, content string) string {
	t.Helper()
	tempFile, err := os.CreateTemp(t.TempDir(), "test_aliases-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return tempFile.Name()
}

func TestNewYAMLProvider(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		wantErr     bool
		expectedErr string
	}{
		{
			name:     "valid file path",
			filePath: "aliases.yaml",
			wantErr:  false,
		},
		{
			name:        "empty file path",
			filePath:    "",
			wantErr:     true,
			expectedErr: "YAML file path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewYAMLProvider(tt.filePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewYAMLProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err.Error() != tt.expectedErr {
					t.Errorf("NewYAMLProvider() error = %q, want %q", err.Error(), tt.expectedErr)
				}
				if provider != nil {
					t.Errorf("NewYAMLProvider() expected nil provider on error, got %v", provider)
				}
			} else {
				if provider == nil {
					t.Errorf("NewYAMLProvider() expected non-nil provider, got nil")
				}
				yp, ok := provider.(*YAMLProvider)
				if !ok {
					t.Errorf("NewYAMLProvider() did not return a *YAMLProvider")
				}
				if yp.filePath != tt.filePath {
					t.Errorf("NewYAMLProvider() filePath = %q, want %q", yp.filePath, tt.filePath)
				}
			}
		})
	}
}

func TestYAMLProvider_GetPredefinedAliases(t *testing.T) {
	validAliasesContent := `
- command: g
  alias: git
- command: k
  alias: kubectl
`
	expectedValidAliases := []alias.Alias{
		{Name: "git", Command: "g"},
		{Name: "kubectl", Command: "k"},
	}

	emptyListContent := `[]`
	emptyFileContent := ``
	malformedContent := `
- name: g
  command: git
  description: git alias
  invalid_field: oops
`
	invalidYAMLContent := `name: g command: git` // Not valid YAML list structure

	tests := []struct {
		name           string
		setupFile      func(t *testing.T) string // Returns the path to the file
		wantAliases    []alias.Alias
		wantErr        bool
		wantErrorMsg   string
		checkErrorType func(err error) bool // Optional: for specific error types like os.IsNotExist
	}{
		{
			name: "file does not exist",
			setupFile: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "non_existent.yaml")
			},
			wantAliases: []alias.Alias{}, // Expect empty slice, not nil
			wantErr:     false,
		},
		{
			name: "empty YAML file (empty content)",
			setupFile: func(t *testing.T) string {
				return createTempYAMLFile(t, emptyFileContent)
			},
			wantAliases: []alias.Alias{}, // Expect empty slice, not nil
			wantErr:     false,
		},
		{
			name: "empty YAML list",
			setupFile: func(t *testing.T) string {
				return createTempYAMLFile(t, emptyListContent)
			},
			wantAliases: []alias.Alias{}, // Expect empty slice, not nil
			wantErr:     false,
		},
		{
			name: "valid aliases file",
			setupFile: func(t *testing.T) string {
				return createTempYAMLFile(t, validAliasesContent)
			},
			wantAliases: expectedValidAliases,
			wantErr:     false,
		},
		{
			name: "malformed YAML content (extra field)",
			setupFile: func(t *testing.T) string {
				return createTempYAMLFile(t, malformedContent)
			},
			wantAliases:  nil,
			wantErr:      true,
			wantErrorMsg: "failed to unmarshal predefined aliases",
		},
		{
			name: "invalid YAML structure",
			setupFile: func(t *testing.T) string {
				return createTempYAMLFile(t, invalidYAMLContent)
			},
			wantAliases:  nil,
			wantErr:      true,
			wantErrorMsg: "failed to unmarshal predefined aliases",
		},
		{
			name: "file is a directory",
			setupFile: func(t *testing.T) string {
				dirPath := filepath.Join(t.TempDir(), "iamadirectory.yaml")
				err := os.Mkdir(dirPath, 0755)
				if err != nil {
					t.Fatalf("Failed to create directory for test: %v", err)
				}
				return dirPath
			},
			wantAliases:  nil,
			wantErr:      true,
			wantErrorMsg: "failed to read predefined aliases file", // os.ReadFile will error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile(t)
			provider, _ := NewYAMLProvider(filePath) // Assume NewYAMLProvider is correct from previous tests

			aliases, err := provider.GetPredefinedAliases()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPredefinedAliases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrorMsg) {
					t.Errorf("GetPredefinedAliases() error = %q, want to contain %q", err.Error(), tt.wantErrorMsg)
				}
			} else {
				if tt.wantAliases == nil && aliases != nil && len(aliases) == 0 {
				} else if !reflect.DeepEqual(aliases, tt.wantAliases) {
					t.Errorf("GetPredefinedAliases() aliases = %v, want %v", aliases, tt.wantAliases)
				}
			}

			if tt.checkErrorType != nil {
				if !tt.checkErrorType(err) {
					t.Errorf("GetPredefinedAliases() error type check failed for error: %v", err)
				}
			}
		})
	}
}
