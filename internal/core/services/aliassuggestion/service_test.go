package aliassuggestion

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/domain/history"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
	"github.com/AntonioJCosta/nicksh/internal/core/testutil"
)

// Helper to sort alias slices for stable comparison
func sortAliases(aliases []alias.Alias) {
	sort.Slice(aliases, func(i, j int) bool {
		return aliases[i].Name < aliases[j].Name
	})
}

func TestNewService(t *testing.T) {
	mockHP := &testutil.MockHistoryProvider{}
	mockAG := &testutil.MockAliasGenerator{}
	mockSCA := &testutil.MockShellConfigAccessor{}
	mockPAP := &testutil.MockPredefinedAliasProvider{}

	t.Run("success with all providers", func(t *testing.T) {
		svc := NewService(mockHP, mockAG, mockSCA, mockPAP)
		if svc == nil {
			t.Fatal("NewService() returned nil, expected a service instance")
		}
	})

	t.Run("success with nil predefinedAliasProvider", func(t *testing.T) {
		svc := NewService(mockHP, mockAG, mockSCA, nil)
		if svc == nil {
			t.Fatal("NewService() returned nil, expected a service instance")
		}
	})

	tests := []struct {
		name                string
		hp                  ports.HistoryProvider
		ag                  ports.AliasGenerator
		sc                  ports.ShellConfigAccessor
		pap                 ports.PredefinedAliasProvider // Added for completeness, though not causing panic if nil
		shouldPanic         bool
		expectedPanicDetail string
	}{
		{"nil historyProvider", nil, mockAG, mockSCA, nil, true, "historyProvider cannot be nil"},
		{"nil aliasGenerator", mockHP, nil, mockSCA, nil, true, "aliasGenerator cannot be nil"},
		{"nil shellConfig", mockHP, mockAG, nil, nil, true, "shellConfig cannot be nil"},
		{"all non-nil providers", mockHP, mockAG, mockSCA, mockPAP, false, ""},
		{"nil predefinedAliasProvider (allowed)", mockHP, mockAG, mockSCA, nil, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.shouldPanic {
					if r == nil {
						t.Errorf("NewService did not panic as expected")
					} else {
						// Check if the panic message matches.
						// Recover returns interface{}, so we need type assertion.
						panicMsg, ok := r.(string)
						if !ok {
							t.Errorf("Panic recovery value is not a string: %T, value: %v", r, r)
						} else if panicMsg != tt.expectedPanicDetail {
							t.Errorf("NewService panicked with wrong message. Got '%s', want '%s'", panicMsg, tt.expectedPanicDetail)
						}
					}
				} else if r != nil {
					t.Errorf("NewService panicked unexpectedly: %v", r)
				}
			}()
			_ = NewService(tt.hp, tt.ag, tt.sc, tt.pap)
		})
	}
}

func TestService_GetFilteredPredefinedAliases(t *testing.T) {
	mockSCA := &testutil.MockShellConfigAccessor{} // Not directly used by GetFilteredPredefinedAliases, but needed for service
	mockHP := &testutil.MockHistoryProvider{}      // Not directly used, but needed for service

	predefinedAliases := []alias.Alias{
		{Name: "pa1", Command: "predefined command 1"},
		{Name: "pa2", Command: "predefined command 2"},
		{Name: "existing", Command: "predefined but exists"}, // Will conflict
	}
	existingShellAliases := map[string]string{
		"existing": "shell command",
		"sh1":      "shell command 1",
	}

	tests := []struct {
		name                  string
		currentShellAliases   map[string]string
		setupMocks            func(ag *testutil.MockAliasGenerator, pap *testutil.MockPredefinedAliasProvider)
		pap                   ports.PredefinedAliasProvider
		wantValidAliases      []alias.Alias
		wantAllLoadedAliases  []alias.Alias
		wantErr               bool
		expectedErrorContains string
	}{
		{
			name:                 "predefinedAliasProvider is nil",
			currentShellAliases:  existingShellAliases,
			pap:                  nil,
			setupMocks:           func(ag *testutil.MockAliasGenerator, pap *testutil.MockPredefinedAliasProvider) {},
			wantValidAliases:     []alias.Alias{},
			wantAllLoadedAliases: []alias.Alias{},
			wantErr:              false,
		},
		{
			name:                "predefined aliases load error (current behavior: no error propagated)",
			currentShellAliases: existingShellAliases,
			pap:                 &testutil.MockPredefinedAliasProvider{},
			setupMocks: func(ag *testutil.MockAliasGenerator, pap *testutil.MockPredefinedAliasProvider) {
				if pap != nil {
					pap.GetPredefinedAliasesFunc = func() ([]alias.Alias, error) {
						return nil, errors.New("failed to load predefined")
					}
				}
			},
			wantValidAliases:     []alias.Alias{},
			wantAllLoadedAliases: []alias.Alias{}, // Because loadAndFilterPredefined returns empty on error
			wantErr:              false,           // GetFilteredPredefinedAliases doesn't return error for this case
		},
		{
			name:                "predefined aliases loaded, some valid, some conflict",
			currentShellAliases: existingShellAliases,
			pap:                 &testutil.MockPredefinedAliasProvider{},
			setupMocks: func(ag *testutil.MockAliasGenerator, pap *testutil.MockPredefinedAliasProvider) {
				if pap != nil {
					pap.GetPredefinedAliasesFunc = func() ([]alias.Alias, error) {
						return predefinedAliases, nil
					}
				}
				ag.IsValidAliasNameFunc = func(name string, existing map[string]string) bool {
					_, existsInShell := existing[name]
					return !existsInShell // Simple mock: valid if not in shell aliases
				}
			},
			wantValidAliases: []alias.Alias{ // "existing" is filtered out
				{Name: "pa1", Command: "predefined command 1"},
				{Name: "pa2", Command: "predefined command 2"},
			},
			wantAllLoadedAliases: predefinedAliases,
			wantErr:              false,
		},
		{
			name:                "predefined aliases loaded, all valid",
			currentShellAliases: map[string]string{"other": "cmd"}, // No conflicts
			pap:                 &testutil.MockPredefinedAliasProvider{},
			setupMocks: func(ag *testutil.MockAliasGenerator, pap *testutil.MockPredefinedAliasProvider) {
				if pap != nil {
					pap.GetPredefinedAliasesFunc = func() ([]alias.Alias, error) {
						return predefinedAliases, nil
					}
				}
				ag.IsValidAliasNameFunc = func(name string, existing map[string]string) bool {
					return true // All are valid
				}
			},
			wantValidAliases:     predefinedAliases,
			wantAllLoadedAliases: predefinedAliases,
			wantErr:              false,
		},
		{
			name:                "no predefined aliases loaded from provider",
			currentShellAliases: existingShellAliases,
			pap:                 &testutil.MockPredefinedAliasProvider{},
			setupMocks: func(ag *testutil.MockAliasGenerator, pap *testutil.MockPredefinedAliasProvider) {
				if pap != nil {
					pap.GetPredefinedAliasesFunc = func() ([]alias.Alias, error) {
						return []alias.Alias{}, nil
					}
				}
			},
			wantValidAliases:     []alias.Alias{},
			wantAllLoadedAliases: []alias.Alias{},
			wantErr:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for each test
			currentAG := &testutil.MockAliasGenerator{}
			// currentPAPInterface := tt.pap // This is ports.PredefinedAliasProvider
			var papToSetup *testutil.MockPredefinedAliasProvider
			if p, ok := tt.pap.(*testutil.MockPredefinedAliasProvider); ok {
				papToSetup = p
			}

			if tt.setupMocks != nil {
				tt.setupMocks(currentAG, papToSetup)
			}

			svc := NewService(mockHP, currentAG, mockSCA, tt.pap) // Pass original tt.pap (interface)
			valid, all, err := svc.GetFilteredPredefinedAliases(tt.currentShellAliases)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetFilteredPredefinedAliases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.expectedErrorContains) {
				t.Errorf("GetFilteredPredefinedAliases() error = %q, want error containing %q", err.Error(), tt.expectedErrorContains)
			}

			sortAliases(valid)
			sortAliases(tt.wantValidAliases)
			if !reflect.DeepEqual(valid, tt.wantValidAliases) {
				t.Errorf("GetFilteredPredefinedAliases() valid aliases = %v, want %v", valid, tt.wantValidAliases)
			}

			sortAliases(all)
			sortAliases(tt.wantAllLoadedAliases)
			if !reflect.DeepEqual(all, tt.wantAllLoadedAliases) {
				t.Errorf("GetFilteredPredefinedAliases() all loaded aliases = %v, want %v", all, tt.wantAllLoadedAliases)
			}
		})
	}
}

func TestService_GetSuggestions(t *testing.T) {
	minFreq, scanLimit, outputLimit := 3, 100, 10

	defaultExistingShellAliases := map[string]string{"ll": "ls -la"}

	tests := []struct {
		name                  string
		setupMocks            func(hp *testutil.MockHistoryProvider, ag *testutil.MockAliasGenerator, sc *testutil.MockShellConfigAccessor, pap *testutil.MockPredefinedAliasProvider)
		pap                   ports.PredefinedAliasProvider
		wantResult            ports.SuggestionResult
		wantErr               bool
		expectedErrorContains string
	}{

		{
			name: "error from GetExistingAliases",
			pap:  nil,
			setupMocks: func(hp *testutil.MockHistoryProvider, ag *testutil.MockAliasGenerator, sc *testutil.MockShellConfigAccessor, pap *testutil.MockPredefinedAliasProvider) {
				sc.GetExistingAliasesFunc = func() (map[string]string, error) { return nil, errors.New("shell config access error") }
			},
			wantErr:               true,
			expectedErrorContains: "failed to get existing aliases",
		},
		{
			name: "error from GetCommandFrequencies",
			pap:  nil,
			setupMocks: func(hp *testutil.MockHistoryProvider, ag *testutil.MockAliasGenerator, sc *testutil.MockShellConfigAccessor, pap *testutil.MockPredefinedAliasProvider) {
				sc.GetExistingAliasesFunc = func() (map[string]string, error) { return defaultExistingShellAliases, nil }
				hp.GetCommandFrequenciesFunc = func(sl, ol int) ([]history.CommandFrequency, error) { return nil, errors.New("history provider error") }
			},
			wantErr:               true,
			expectedErrorContains: "failed to get command frequencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHP := &testutil.MockHistoryProvider{}
			mockAG := &testutil.MockAliasGenerator{}
			mockSC := &testutil.MockShellConfigAccessor{}
			var concretePAP *testutil.MockPredefinedAliasProvider
			if p, ok := tt.pap.(*testutil.MockPredefinedAliasProvider); ok {
				concretePAP = p
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockHP, mockAG, mockSC, concretePAP)
			}

			svc := NewService(mockHP, mockAG, mockSC, tt.pap) // Use tt.pap (interface) for NewService
			result, err := svc.GetSuggestions(minFreq, scanLimit, outputLimit)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSuggestions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.expectedErrorContains) {
				t.Errorf("GetSuggestions() error = %q, want error containing %q", err.Error(), tt.expectedErrorContains)
			}
			if !tt.wantErr {
				sortAliases(result.Suggestions)
				sortAliases(tt.wantResult.Suggestions)
				if !reflect.DeepEqual(result.Suggestions, tt.wantResult.Suggestions) {
					t.Errorf("GetSuggestions() suggestions = %v, want %v", result.Suggestions, tt.wantResult.Suggestions)
				}
				if result.SourceDetails != tt.wantResult.SourceDetails {
					t.Errorf("GetSuggestions() sourceDetails = %q, want %q", result.SourceDetails, tt.wantResult.SourceDetails)
				}
			}
		})
	}
}

func TestService_GetSuggestionContextDetails(t *testing.T) {
	mockAG := &testutil.MockAliasGenerator{}      // Needed for NewService
	mockSC := &testutil.MockShellConfigAccessor{} // Needed for NewService
	historySourceID := "File: /path/to/zsh_history"

	tests := []struct {
		name                  string
		setupMocks            func(hp *testutil.MockHistoryProvider, pap *testutil.MockPredefinedAliasProvider)
		pap                   ports.PredefinedAliasProvider
		wantDetails           string
		wantErr               bool
		expectedErrorContains string // Not expecting errors from this method based on current impl
	}{
		{
			name: "success - no predefined provider",
			pap:  nil,
			setupMocks: func(hp *testutil.MockHistoryProvider, pap *testutil.MockPredefinedAliasProvider) {
				hp.GetSourceIdentifierFunc = func() string { return historySourceID }
			},
			wantDetails: historySourceID,
			wantErr:     false,
		},
		{
			name: "success - predefined provider configured and loads successfully",
			pap:  &testutil.MockPredefinedAliasProvider{},
			setupMocks: func(hp *testutil.MockHistoryProvider, pap *testutil.MockPredefinedAliasProvider) {
				hp.GetSourceIdentifierFunc = func() string { return historySourceID }
				if pap != nil {
					pap.GetPredefinedAliasesFunc = func() ([]alias.Alias, error) {
						return []alias.Alias{{Name: "p", Command: "c"}}, nil
					}
				}
			},
			wantDetails: historySourceID + " (predefined aliases are configured and loadable)",
			wantErr:     false,
		},
		{
			name: "success - predefined provider configured but fails to load",
			pap:  &testutil.MockPredefinedAliasProvider{},
			setupMocks: func(hp *testutil.MockHistoryProvider, pap *testutil.MockPredefinedAliasProvider) {
				hp.GetSourceIdentifierFunc = func() string { return historySourceID }
				if pap != nil {
					pap.GetPredefinedAliasesFunc = func() ([]alias.Alias, error) {
						return nil, errors.New("load failed")
					}
				}
			},
			wantDetails: fmt.Sprintf("%s (predefined aliases configured but failed to load: %v)", historySourceID, errors.New("load failed")),
			wantErr:     false, // The error from GetPredefinedAliases is part of the details string, not returned by GetSuggestionContextDetails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHP := &testutil.MockHistoryProvider{}
			var concretePAP *testutil.MockPredefinedAliasProvider
			if p, ok := tt.pap.(*testutil.MockPredefinedAliasProvider); ok {
				concretePAP = p
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockHP, concretePAP)
			}

			svc := NewService(mockHP, mockAG, mockSC, tt.pap) // Use tt.pap (interface) for NewService
			details, err := svc.GetSuggestionContextDetails()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSuggestionContextDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// According to current implementation, GetSuggestionContextDetails does not return errors.
			// If it could, this check would be relevant:
			// if tt.wantErr && !strings.Contains(err.Error(), tt.expectedErrorContains) {
			// 	t.Errorf("GetSuggestionContextDetails() error = %q, want error containing %q", err.Error(), tt.expectedErrorContains)
			// }
			if !tt.wantErr && details != tt.wantDetails {
				t.Errorf("GetSuggestionContextDetails() = %q, want %q", details, tt.wantDetails)
			}
		})
	}
}
