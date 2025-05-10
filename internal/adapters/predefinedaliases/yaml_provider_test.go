package predefinedaliases

import (
	"reflect"
	"strings"
	"testing"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
)

func TestNewYAMLProvider(t *testing.T) {
	provider, err := NewYAMLProvider()

	if err != nil {
		t.Errorf("NewYAMLProvider() unexpected error = %v", err)
	}
	if provider == nil {
		t.Errorf("NewYAMLProvider() expected non-nil provider, got nil")
	}
	if _, ok := provider.(*YAMLProvider); !ok {
		t.Errorf("NewYAMLProvider() did not return a *YAMLProvider, got %T", provider)
	}
}

func TestYAMLProvider_GetPredefinedAliases(t *testing.T) {
	// Ensure these YAML structures match the fields in alias.Alias (name, command, description)
	validAliasesYAML := `
- command: git
  alias: g
- command: kubectl
  alias: k
`
	expectedValidAliases := []alias.Alias{
		{Name: "g", Command: "git"},
		{Name: "k", Command: "kubectl"},
	}

	emptyListYAML := `[]`
	emptyContentYAML := `` // Represents an empty embedded file (0 bytes)
	malformedContentWithExtraFieldYAML := `
- name: g
  command: git
  invalid_field: "this should cause an error with KnownFields(true)" # Explains the malformed nature
`
	invalidYAMLStructure := `name: g command: git` // Not a valid YAML list

	// Store the original value of the package-level embeddedPredefinedAliases.
	// This is important for test isolation, especially if tests run in an environment
	// where it might have been populated by a previous build step.
	originalEmbeddedData := embeddedPredefinedAliases

	tests := []struct {
		name                string
		contentToEmbed      []byte
		wantAliases         []alias.Alias
		wantErr             bool
		wantErrorMsgSnippet string // A snippet of the expected error message if wantErr is true
	}{
		{
			name:           "embedded content is nil (simulates no file linked or empty at compile time)",
			contentToEmbed: nil, // This will result in len(embeddedPredefinedAliases) == 0
			wantAliases:    []alias.Alias{},
			wantErr:        false,
		},
		{
			name:           "embedded content is empty string (0 bytes)",
			contentToEmbed: []byte(emptyContentYAML),
			wantAliases:    []alias.Alias{},
			wantErr:        false,
		},
		{
			name:           "embedded content is an empty YAML list",
			contentToEmbed: []byte(emptyListYAML),
			wantAliases:    []alias.Alias{},
			wantErr:        false,
		},
		{
			name:           "valid aliases embedded",
			contentToEmbed: []byte(validAliasesYAML),
			wantAliases:    expectedValidAliases,
			wantErr:        false,
		},
		{
			name:                "malformed YAML content (extra field with KnownFields=true)",
			contentToEmbed:      []byte(malformedContentWithExtraFieldYAML),
			wantAliases:         nil, // On error, expect nil aliases
			wantErr:             true,
			wantErrorMsgSnippet: "failed to unmarshal embedded predefined aliases",
		},
		{
			name:                "invalid YAML structure (not a list)",
			contentToEmbed:      []byte(invalidYAMLStructure),
			wantAliases:         nil, // On error, expect nil aliases
			wantErr:             true,
			wantErrorMsgSnippet: "failed to unmarshal embedded predefined aliases",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the package-level variable for this specific test case
			embeddedPredefinedAliases = tt.contentToEmbed
			// Ensure the original value is restored after this test case finishes for test isolation
			t.Cleanup(func() {
				embeddedPredefinedAliases = originalEmbeddedData
			})

			provider, err := NewYAMLProvider()
			if err != nil {
				t.Fatalf("NewYAMLProvider() failed unexpectedly: %v", err)
			}

			aliases, err := provider.GetPredefinedAliases()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPredefinedAliases() error = %v, wantErr %v", err, tt.wantErr)
				return // Important to return if error expectation is mismatched
			}

			if tt.wantErr {
				if tt.wantErrorMsgSnippet == "" {
					t.Errorf("GetPredefinedAliases() wantErrorMsgSnippet is empty for an error case")
				} else if !strings.Contains(err.Error(), tt.wantErrorMsgSnippet) {
					t.Errorf("GetPredefinedAliases() error = %q, want error to contain %q", err.Error(), tt.wantErrorMsgSnippet)
				}
				// When an error is expected, the returned aliases should ideally be nil
				if aliases != nil {
					t.Errorf("GetPredefinedAliases() expected nil aliases on error, got %#v", aliases)
				}
			}

			// reflect.DeepEqual handles nil slices and empty slices correctly.
			if !reflect.DeepEqual(aliases, tt.wantAliases) {
				t.Errorf("GetPredefinedAliases() aliases = %#v, want %#v", aliases, tt.wantAliases)
			}
		})
	}
}
