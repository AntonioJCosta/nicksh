package testutil

import (
	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/domain/history"
)

// MockAliasGenerator is a mock implementation of ports.AliasGenerator.
type MockAliasGenerator struct {
	GenerateSuggestionsFunc func(frequencies []history.CommandFrequency, existingAliases map[string]string, minFrequency int) []alias.Alias
	IsValidAliasNameFunc    func(name string, existingAliases map[string]string) bool // Added field for the new method
}

func (m *MockAliasGenerator) GenerateSuggestions(frequencies []history.CommandFrequency, existingAliases map[string]string, minFrequency int) []alias.Alias {
	if m.GenerateSuggestionsFunc != nil {
		return m.GenerateSuggestionsFunc(frequencies, existingAliases, minFrequency)
	}
	// Consider if a panic is more appropriate: panic("MockAliasGenerator: GenerateSuggestionsFunc not implemented")
	return []alias.Alias{} // Return empty slice if not implemented
}

// IsValidAliasName implements the ports.AliasGenerator interface.
func (m *MockAliasGenerator) IsValidAliasName(name string, existingAliases map[string]string) bool {
	if m.IsValidAliasNameFunc != nil {
		return m.IsValidAliasNameFunc(name, existingAliases)
	}
	// Default behavior if not implemented: consider what makes sense for tests.
	// Returning true might be a safe default, or panicking if the behavior is critical.
	// panic("MockAliasGenerator: IsValidAliasNameFunc not implemented")
	return true // Default to true if not implemented
}

// In testutil/mocks.go or similar
type MockPredefinedAliasProvider struct {
	GetPredefinedAliasesFunc func() ([]alias.Alias, error)
}

func (m *MockPredefinedAliasProvider) GetPredefinedAliases() ([]alias.Alias, error) {
	if m.GetPredefinedAliasesFunc != nil {
		return m.GetPredefinedAliasesFunc()
	}
	return nil, nil // Default behavior
}
