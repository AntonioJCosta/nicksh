package ports

import (
	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/domain/history" // Or your types package
)

/*
AliasGenerator defines the contract for a service that generates alias suggestions.
This is a driven port, representing a domain capability.
*/
type AliasGenerator interface {
	GenerateSuggestions(
		commands []history.CommandFrequency,
		existingAliases map[string]string, // Aliases to avoid generating
		minFrequency int,
	) []alias.Alias

	// IsValidAliasName checks if a given name is valid according to general system rules
	// (e.g., not a system command, valid characters, not in the provided existing map).
	// It takes the name to check and a map of already existing/forbidden names.
	IsValidAliasName(nameToCheck string, existingAliases map[string]string) bool
}
