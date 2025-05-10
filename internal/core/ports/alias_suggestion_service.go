package ports

import "github.com/AntonioJCosta/nicksh/internal/core/domain/alias"

// SuggestionResult holds the suggestions and any relevant metadata.
type SuggestionResult struct {
	Suggestions   []alias.Alias
	SourceDetails string
}

// AliasSuggestionService defines the contract for generating alias suggestions.
type AliasSuggestionService interface {
	GetSuggestions(minFrequency, scanLimit, outputLimit int) (SuggestionResult, error)
	GetSuggestionContextDetails() (string, error)
	// GetFilteredPredefinedAliases loads predefined aliases and filters them based on validity
	// and conflicts with the provided currentShellAliases.
	// It returns the list of valid aliases, the list of all aliases originally loaded, and any error encountered.
	GetFilteredPredefinedAliases(currentShellAliases map[string]string) (validAliases []alias.Alias, allLoadedAliases []alias.Alias, err error)
}
