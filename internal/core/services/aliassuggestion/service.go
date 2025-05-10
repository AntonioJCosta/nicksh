package aliassuggestion

import (
	"fmt"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
)

type service struct {
	historyProvider         ports.HistoryProvider
	aliasGenerator          ports.AliasGenerator
	shellConfig             ports.ShellConfigAccessor
	predefinedAliasProvider ports.PredefinedAliasProvider // Can be nil if no predefined aliases are configured.
}

// NewService creates a new alias suggestion service.
// It panics if historyProvider, aliasGenerator, or shellConfigAccessor are nil.
// predefinedAliasProvider can be nil if not used.
func NewService(
	hp ports.HistoryProvider,
	ag ports.AliasGenerator,
	sc ports.ShellConfigAccessor,
	pap ports.PredefinedAliasProvider,
) ports.AliasSuggestionService {
	if hp == nil {
		panic("historyProvider cannot be nil")
	}
	if ag == nil {
		panic("aliasGenerator cannot be nil")
	}
	if sc == nil {
		panic("shellConfig cannot be nil")
	}
	// predefinedAliasProvider is allowed to be nil.
	return &service{
		historyProvider:         hp,
		aliasGenerator:          ag,
		shellConfig:             sc,
		predefinedAliasProvider: pap,
	}
}

// GetFilteredPredefinedAliases loads predefined aliases and filters them against existing shell aliases.
// It returns the list of valid predefined aliases and the original list of all loaded predefined aliases.
// Returns an error if the predefined alias provider or alias generator is not configured.
func (s *service) GetFilteredPredefinedAliases(currentShellAliases map[string]string) ([]alias.Alias, []alias.Alias, error) {
	if s.predefinedAliasProvider == nil {
		// If no provider is configured, there are no predefined aliases to process.
		return []alias.Alias{}, []alias.Alias{}, nil
	}
	if s.aliasGenerator == nil {
		// Should not happen if NewService is used, but defensive check.
		return []alias.Alias{}, []alias.Alias{}, fmt.Errorf("alias generator is not configured")
	}

	validAliases, allLoaded, err := s.loadAndFilterPredefined(currentShellAliases)
	if err != nil {
		// loadAndFilterPredefined currently returns nil error even on load issues,
		// but if it were to change, this would propagate the error.
		return nil, nil, fmt.Errorf("error processing predefined aliases: %w", err)
	}
	return validAliases, allLoaded, nil
}
func (s *service) GetSuggestions(minFrequency, scanLimit, outputLimit int) (ports.SuggestionResult, error) {
	var result ports.SuggestionResult

	existingShellAliases, err := s.shellConfig.GetExistingAliases()
	if err != nil {
		return result, fmt.Errorf("failed to get existing aliases for suggestion generation: %w", err)
	}

	var validPredefined, allLoadedPredefined []alias.Alias
	if s.predefinedAliasProvider != nil {
		// Load and filter predefined aliases primarily to know their names for conflict avoidance.
		validPredefined, allLoadedPredefined, err = s.loadAndFilterPredefined(existingShellAliases)
		if err != nil {
			// loadAndFilterPredefined currently returns nil error even on load issues.
			// If this changes, this error handling will be relevant.
			// A warning might be logged inside loadAndFilterPredefined.
		}
	}

	// Build a map of names that should not be used for dynamic alias generation.
	// This includes names from existing shell aliases AND valid predefined aliases.
	forbiddenNamesForDynamicGen := s.buildForbiddenNamesMap(existingShellAliases, validPredefined)

	frequencies, err := s.historyProvider.GetCommandFrequencies(scanLimit, outputLimit)
	if err != nil {
		return result, fmt.Errorf("failed to get command frequencies: %w", err)
	}

	dynamicSuggestions := s.aliasGenerator.GenerateSuggestions(frequencies, forbiddenNamesForDynamicGen, minFrequency)

	// Pass an empty slice for predefined aliases to combineSuggestions,
	// ensuring only dynamic suggestions are processed for the final list.
	// The combineSuggestions method will handle de-duplication of dynamic suggestions if any (though ideally none).
	result.Suggestions = s.combineSuggestions([]alias.Alias{}, dynamicSuggestions)

	result.SourceDetails = s.historyProvider.GetSourceIdentifier()
	result.SourceDetails += " (suggestions from command history" // Base part of the message

	if s.predefinedAliasProvider != nil { // Check if predefined aliases are configured at all
		if len(allLoadedPredefined) > 0 {
			// This implies predefined aliases were loaded and thus considered for conflict avoidance.
			result.SourceDetails += "; predefined aliases considered for conflict avoidance"
		} else {
			// Predefined provider is configured, but no aliases were loaded.
			result.SourceDetails += "; predefined aliases configured but none loaded/found for conflict avoidance"
		}
	}
	result.SourceDetails += ")" // Close the parenthesis

	return result, nil
}

// GetSuggestionContextDetails provides details about the sources used for suggestions.
func (s *service) GetSuggestionContextDetails() (string, error) {
	details := s.historyProvider.GetSourceIdentifier()
	if s.predefinedAliasProvider != nil {
		// Check if predefined aliases are configured and attempt to load them to confirm.
		// This doesn't need the actual aliases, just confirmation of their status.
		_, loadErr := s.predefinedAliasProvider.GetPredefinedAliases()
		if loadErr == nil {
			details += " (predefined aliases are configured and loadable)"
		} else {
			details += fmt.Sprintf(" (predefined aliases configured but failed to load: %v)", loadErr)
		}
	}
	return details, nil
}
