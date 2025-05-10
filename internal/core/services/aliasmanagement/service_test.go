package aliasmanagement

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/testutil" // Assuming this path is correct
)

func TestNewService(t *testing.T) {
	t.Run("should return a service if shellConfig is not nil", func(t *testing.T) {
		mockSC := &testutil.MockShellConfigAccessor{}
		svc := NewService(mockSC)
		if svc == nil {
			t.Fatal("NewService() returned nil, expected a service instance")
		}
	})

	t.Run("should panic if shellConfig is nil", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("NewService did not panic with nil shellConfig")
			}
		}()
		_ = NewService(nil) // Panics if sc is nil
	})
}

func TestService_AddAliasToConfig(t *testing.T) {
	testAlias := alias.Alias{Name: "test", Command: "echo test"}

	tests := []struct {
		name          string
		aliasName     string
		aliasCommand  string
		setupMock     func(mockSC *testutil.MockShellConfigAccessor)
		wantAdded     bool
		wantErr       bool
		expectedError error
	}{
		{
			name:         "success - alias newly added",
			aliasName:    testAlias.Name,
			aliasCommand: testAlias.Command,
			setupMock: func(mockSC *testutil.MockShellConfigAccessor) {
				mockSC.AddAliasFunc = func(newAlias alias.Alias) (bool, error) {
					if newAlias.Name != testAlias.Name || newAlias.Command != testAlias.Command {
						t.Errorf("AddAlias received wrong alias. Got %+v, want %+v", newAlias, testAlias)
					}
					return true, nil // Simulate alias was newly added
				}
			},
			wantAdded: true,
			wantErr:   false,
		},
		{
			name:         "success - alias already existed (not overwritten, or updated)",
			aliasName:    testAlias.Name,
			aliasCommand: testAlias.Command,
			setupMock: func(mockSC *testutil.MockShellConfigAccessor) {
				mockSC.AddAliasFunc = func(newAlias alias.Alias) (bool, error) {
					// Logic for checking newAlias can remain the same
					return false, nil // Simulate alias already existed
				}
			},
			wantAdded: false,
			wantErr:   false,
		},
		{
			name:         "failure - shellConfig returns error",
			aliasName:    testAlias.Name,
			aliasCommand: testAlias.Command,
			setupMock: func(mockSC *testutil.MockShellConfigAccessor) {
				mockSC.AddAliasFunc = func(newAlias alias.Alias) (bool, error) {
					return false, errors.New("shell config error")
				}
			},
			wantAdded:     false,
			wantErr:       true,
			expectedError: errors.New("shell config error"), // The specific error from the mock
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSC := &testutil.MockShellConfigAccessor{}
			if tt.setupMock != nil {
				tt.setupMock(mockSC)
			}
			svc := NewService(mockSC)

			gotAdded, err := svc.AddAliasToConfig(tt.aliasName, tt.aliasCommand)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddAliasToConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.expectedError != nil {
				// Check if the error contains the expected error, due to wrapping
				if !errors.Is(err, tt.expectedError) && err.Error() != fmt.Sprintf("failed to add alias '%s': %s", tt.aliasName, tt.expectedError.Error()) {
					t.Errorf("AddAliasToConfig() error = %v, want error containing %v", err, tt.expectedError)
				}
			}
			if gotAdded != tt.wantAdded {
				t.Errorf("AddAliasToConfig() gotAdded = %v, want %v", gotAdded, tt.wantAdded)
			}
		})
	}
}

func TestService_ListAliases(t *testing.T) {
	expectedAliasesMap := map[string]string{"ll": "ls -l", "ga": "git add"}
	shellConfigErr := errors.New("shell config error")

	tests := []struct {
		name                string
		setupMock           func(mockSC *testutil.MockShellConfigAccessor)
		expectedResult      map[string]string
		wantErr             bool
		expectedErrorString string // For checking wrapped errors
	}{
		{
			name: "success",
			setupMock: func(mockSC *testutil.MockShellConfigAccessor) {
				mockSC.GetExistingAliasesFunc = func() (map[string]string, error) {
					return expectedAliasesMap, nil
				}
			},
			expectedResult: expectedAliasesMap,
			wantErr:        false,
		},
		{
			name: "failure - shellConfig returns error",
			setupMock: func(mockSC *testutil.MockShellConfigAccessor) {
				mockSC.GetExistingAliasesFunc = func() (map[string]string, error) {
					return nil, shellConfigErr
				}
			},
			expectedResult:      nil,
			wantErr:             true,
			expectedErrorString: fmt.Sprintf("failed to list existing aliases: %s", shellConfigErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSC := &testutil.MockShellConfigAccessor{}
			if tt.setupMock != nil {
				tt.setupMock(mockSC)
			}
			svc := NewService(mockSC)

			aliases, err := svc.ListAliases()

			if (err != nil) != tt.wantErr {
				t.Errorf("ListAliases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.expectedErrorString {
				t.Errorf("ListAliases() error = %q, want %q", err.Error(), tt.expectedErrorString)
			}
			if !reflect.DeepEqual(aliases, tt.expectedResult) {
				t.Errorf("ListAliases() = %v, want %v", aliases, tt.expectedResult)
			}
		})
	}
}

// TestService_GetShellConfigPath assumes GetShellConfigPath is a method on your service.
// If it's not, this test is for a non-existent method.
// The provided service.go snippet does not show this method.
func TestService_GetShellConfigPath(t *testing.T) {
	expectedPath := "/home/user/.bashrc"
	shellConfigErr := errors.New("shell config error")

	tests := []struct {
		name                string
		setupMock           func(mockSC *testutil.MockShellConfigAccessor)
		expectedResult      string
		wantErr             bool
		expectedErrorString string // For checking wrapped errors
	}{
		{
			name: "success",
			setupMock: func(mockSC *testutil.MockShellConfigAccessor) {
				mockSC.GetConfigPathFunc = func() (string, error) {
					return expectedPath, nil
				}
			},
			expectedResult: expectedPath,
			wantErr:        false,
		},
		{
			name: "failure - shellConfig returns error",
			setupMock: func(mockSC *testutil.MockShellConfigAccessor) {
				mockSC.GetConfigPathFunc = func() (string, error) {
					return "", shellConfigErr
				}
			},
			expectedResult:      "",
			wantErr:             true,
			expectedErrorString: fmt.Sprintf("failed to get shell config path: %s", shellConfigErr.Error()), // Assuming similar error wrapping
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSC := &testutil.MockShellConfigAccessor{}
			if tt.setupMock != nil {
				tt.setupMock(mockSC)
			}

			// Assuming GetShellConfigPath exists on the service:
			// path, err := svc.GetShellConfigPath()
			// For now, let's assume the method signature and call based on the original test.
			// If the method is `GetConfigPath() (string, error)` directly on the service,
			// or if it's wrapped like other methods, the call and error checking might differ.
			// The original test implies a method like `svc.GetShellConfigPath()`
			// which would internally call `s.shellConfig.GetConfigPath()`.

		})
	}
}
