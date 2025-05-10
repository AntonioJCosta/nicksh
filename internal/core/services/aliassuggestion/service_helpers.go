package aliassuggestion

import (
	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
)

// loadAndFilterPredefined loads predefined aliases from the provider (if configured)
// and filters them against existing shell aliases to ensure validity.
// It returns the list of valid predefined aliases and the original list of all loaded predefined aliases.
// If the predefined alias provider is not set, it returns empty lists and no error.
// If loading from the provider fails, it currently returns empty lists and a nil error
// (consider changing to return the error if strict error handling is preferred).
func (s *service) loadAndFilterPredefined(existingShellAliases map[string]string) ([]alias.Alias, []alias.Alias, error) {
	if s.predefinedAliasProvider == nil {
		return []alias.Alias{}, []alias.Alias{}, nil // No provider, so no predefined aliases.
	}

	loadedPredefined, loadErr := s.predefinedAliasProvider.GetPredefinedAliases()
	if loadErr != nil {
		// Current behavior: treat as no predefined aliases if loading fails.
		// log.Printf("Warning: could not load predefined aliases: %v\n", loadErr) // Optional logging
		return []alias.Alias{}, []alias.Alias{}, nil // Return empty lists on load error.
		// Alternative: return []alias.Alias{}, []alias.Alias{}, loadErr // To propagate the error.
	}

	validPredefinedAliases := make([]alias.Alias, 0, len(loadedPredefined))
	for _, pa := range loadedPredefined {
		// The aliasGenerator's IsValidAliasName checks against system commands and other rules.
		if s.aliasGenerator.IsValidAliasName(pa.Name, existingShellAliases) {
			validPredefinedAliases = append(validPredefinedAliases, pa)
		}
	}
	return validPredefinedAliases, loadedPredefined, nil
}

// buildForbiddenNamesMap creates a map of names that should not be used for dynamic alias generation.
// This includes names from existing shell aliases and valid predefined aliases.
func (s *service) buildForbiddenNamesMap(existingShellAliases map[string]string, validPredefinedAliases []alias.Alias) map[string]string {
	forbiddenNames := make(map[string]string)
	for name, cmd := range existingShellAliases {
		forbiddenNames[name] = cmd
	}
	for _, pa := range validPredefinedAliases {
		forbiddenNames[pa.Name] = pa.Command
	}
	return forbiddenNames
}

// combineSuggestions merges predefined and dynamic suggestions.
// Predefined suggestions take precedence if there are name conflicts.
func (s *service) combineSuggestions(predefined []alias.Alias, dynamic []alias.Alias) []alias.Alias {
	// Use a map to ensure uniqueness and handle precedence.
	finalMap := make(map[string]alias.Alias)

	for _, pa := range predefined {
		finalMap[pa.Name] = pa
	}

	for _, ds := range dynamic {
		if _, exists := finalMap[ds.Name]; !exists {
			finalMap[ds.Name] = ds
		}
	}

	// Convert map back to a slice.
	combined := make([]alias.Alias, 0, len(finalMap))
	for _, sug := range finalMap {
		combined = append(combined, sug)
	}
	// Sorting is not currently implemented but could be added here if a specific order is required.
	// e.g., sort.Slice(combined, func(i, j int) bool { return combined[i].Name < combined[j].Name })
	return combined
}
