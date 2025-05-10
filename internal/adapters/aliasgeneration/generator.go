package aliasgeneration

import (
	"os/exec"
	"regexp"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/domain/history"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
)

// AliasGenerator generates alias suggestions based on command history.
type AliasGenerator struct {
	analyzer ports.CommandAnalyzer
}

// NewAliasGenerator creates a new AliasGenerator.
func NewAliasGenerator(analyzer ports.CommandAnalyzer) ports.AliasGenerator {
	return &AliasGenerator{analyzer: analyzer}
}

// GenerateSuggestions creates alias suggestions from command frequencies using multiple strategies.
func (g *AliasGenerator) GenerateSuggestions(
	commands []history.CommandFrequency,
	existingAliases map[string]string, // Aliases already defined in the user's environment.
	minFrequency int, // Minimum frequency for a command to be considered.
) []alias.Alias {
	allSuggestions := []alias.Alias{}
	// Tracks names generated in this run to avoid duplicates from different strategies.
	generatedNamesInThisRun := make(map[string]bool)
	// Minimum effective length (non-space characters) for a command to be considered by some strategies.
	const minCommandEffectiveLength = 4

	// Strategy 1: Aliases for "command + first non-flag argument" patterns (e.g., "git pull" -> "gp").
	cmdFirstArgFreq, cmdFirstArgToAnalyzedCmd := g.aggregateForCommandFirstArgStrategy(
		commands,
		minCommandEffectiveLength,
	)
	strategy1Suggestions := g.generateAliasesFromCommandFirstArgAggregation(
		cmdFirstArgFreq,
		cmdFirstArgToAnalyzedCmd,
		minFrequency,
		existingAliases,
		generatedNamesInThisRun,
	)
	allSuggestions = append(allSuggestions, strategy1Suggestions...)

	// Strategy 2: Aliases for exact full command strings (e.g., "git commit -m 'feat: initial'" -> "gcm").
	strategy2Suggestions := g.generateExactFullCommandAliasesStrategy(
		commands,
		minFrequency,
		existingAliases,
		generatedNamesInThisRun,
		minCommandEffectiveLength,
	)
	allSuggestions = append(allSuggestions, strategy2Suggestions...)

	// Future strategies could be added here.
	// e.g., common misspellings, command-only aliases for long commands.

	return allSuggestions
}

// validAliasCharsRegexGenerator ensures generated alias names are alphanumeric.
var validAliasCharsRegexGenerator = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// IsValidAliasName checks if a given name is suitable for use as an alias.
// It verifies length, character set, and conflicts with existing aliases or system commands.
func (g *AliasGenerator) IsValidAliasName(nameToCheck string, existingAliases map[string]string) bool {
	// Rule: Alias must be at least 1 character long.
	if len(nameToCheck) < 1 { // Consider making this minimum length configurable.
		return false
	}
	// Rule: Alias must only contain alphanumeric characters.
	if !validAliasCharsRegexGenerator.MatchString(nameToCheck) {
		return false
	}
	// Rule: Alias must not conflict with existing aliases.
	if _, exists := existingAliases[nameToCheck]; exists {
		return false
	}
	// Rule: Alias must not conflict with system commands.
	if _, err := exec.LookPath(nameToCheck); err == nil {
		// Name corresponds to an executable in PATH, so it's a conflict.
		return false
	}
	return true
}
