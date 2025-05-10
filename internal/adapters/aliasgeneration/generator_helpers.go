package aliasgeneration

import (
	"regexp"
	"strings"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/domain/command"
	"github.com/AntonioJCosta/nicksh/internal/core/domain/history"
)

// isAliasNameValid is a local helper to check if a proposed name exists in the provided map.
func isAliasNameValid(
	proposedName string,
	existingAliases map[string]string,
) bool {
	// Check against existing aliases provided.
	if _, exists := existingAliases[proposedName]; exists {
		return false
	}
	return true
}

// Regex to allow only alphanumeric characters and dots in alias names
var validAliasCharsRegex = regexp.MustCompile(`^[a-zA-Z0-9.]+$`)

/*
isProposedNameValid checks common validation rules for a proposed alias name.

It verifies that the alias is of a minimum length, not the same as the
original command, not already generated in the current suggestion run,
and does not conflict with existing aliases. System command conflict check
is handled by the main IsValidAliasName method.

Example:

	isValid := g.isProposedNameValid("gp", "git", existing, generated)
	// isValid would be true if "gp" is >= 2 chars, not "git", not in generated,
	// and not an existing alias.
*/
func (g *AliasGenerator) isProposedNameValid(
	proposedName string,
	originalCommandName string,
	existingAliases map[string]string,
	generatedNamesInThisRun map[string]bool,
) bool {
	// Rule: Alias must be at least 2 characters long.
	if len(proposedName) < 2 {
		return false
	}
	// Rule: Alias must only contain alphanumeric characters.
	if !validAliasCharsRegex.MatchString(proposedName) {
		return false
	}
	// Rule: Alias should not be the same as the original command name.
	if proposedName == originalCommandName {
		return false
	}
	// Rule: Alias must not have been generated already in the current suggestion run.
	if _, exists := generatedNamesInThisRun[proposedName]; exists {
		return false
	}
	// Rule: Alias must not conflict with existing aliases (checked by local helper).
	return isAliasNameValid(proposedName, existingAliases)
}

/*
generateCommandSubcommandAliasName generates a short alias for a "command + first_main_argument" pattern.
For example, "git pull" might become "gp".
It prioritizes common patterns like "cd .." -> "up".
*/
func (g *AliasGenerator) generateCommandSubcommandAliasName(analyzedCmd command.AnalyzedCommand) string {
	if analyzedCmd.CommandName == "" {
		return ""
	}

	if analyzedCmd.CommandName == "cd" && len(analyzedCmd.PotentialArgs) == 1 && analyzedCmd.PotentialArgs[0] == ".." {
		return "up"
	}

	cmdNameLower := strings.ToLower(analyzedCmd.CommandName)
	aliasInitial := string(cmdNameLower[0])

	if len(analyzedCmd.PotentialArgs) > 0 && len(analyzedCmd.PotentialArgs[0]) > 0 {
		firstArg := analyzedCmd.PotentialArgs[0]
		if strings.Contains(firstArg, "/") {
			parts := strings.Split(firstArg, "/")
			lastPart := parts[len(parts)-1]
			if lastPart != "" && !strings.HasPrefix(lastPart, ".") {
				aliasInitial += string(strings.ToLower(lastPart)[0])
			} else if len(firstArg) > 0 {
				aliasInitial += string(strings.ToLower(firstArg)[0])
			}
		} else if !strings.HasPrefix(firstArg, "-") { // Regular argument
			aliasInitial += string(strings.ToLower(firstArg)[0])
		} else if len(cmdNameLower) > 1 { // Fallback for flags or if arg logic didn't add anything
			aliasInitial = cmdNameLower[:2]
		}
		// If command is 1 char and arg is a flag, aliasInitial remains 1 char.
		// isProposedNameValid will filter if len < 2.
	} else if len(cmdNameLower) > 1 {
		aliasInitial = cmdNameLower[:2]
	}
	return aliasInitial
}

/*
generateExactCommandAliasName generates a specific alias for a full command string.
It attempts to create a concise alias, e.g., "git add ." -> "ga.".
Handles common patterns like "cd .." -> "up".
For complex commands, it might generate a simpler alias based on the command name.
*/
func (g *AliasGenerator) generateExactCommandAliasName(analyzedCmd command.AnalyzedCommand) string {
	if analyzedCmd.CommandName == "" {
		return ""
	}

	if analyzedCmd.CommandName == "cd" && len(analyzedCmd.PotentialArgs) == 1 && analyzedCmd.PotentialArgs[0] == ".." {
		return "up"
	}

	if analyzedCmd.IsComplex {
		if len(analyzedCmd.CommandName) >= 2 {
			return strings.ToLower(analyzedCmd.CommandName[:2])
		}
		return string(strings.ToLower(analyzedCmd.CommandName)[0]) // Likely filtered by length check later.
	}

	nameParts := []string{string(strings.ToLower(analyzedCmd.CommandName)[0])}
	argInitialsCount := 0
	const maxArgInitials = 3 // Max argument initials to include after command initial.

	for _, arg := range analyzedCmd.PotentialArgs {
		if argInitialsCount >= maxArgInitials {
			break
		}
		if len(arg) == 0 {
			continue
		}

		cleanArg := strings.ToLower(arg)
		var partToAdd string

		if strings.HasPrefix(cleanArg, "--") && len(cleanArg) > 2 {
			partToAdd = string(cleanArg[2])
		} else if strings.HasPrefix(cleanArg, "-") && len(cleanArg) > 1 {
			partToAdd = string(cleanArg[1])
		} else if cleanArg == "." {
			partToAdd = "."
		} else if cleanArg == ".." {
			// ".." is typically handled by "up" or CommandSubcommand strategy. Avoid adding to exact names.
			continue
		} else if strings.Contains(cleanArg, "/") { // Handle paths
			pathParts := strings.Split(cleanArg, "/")
			significantPart := ""
			for i := len(pathParts) - 1; i >= 0; i-- {
				if pathParts[i] != "" && !strings.HasPrefix(pathParts[i], ".") {
					significantPart = pathParts[i]
					break
				}
			}
			if significantPart == "" && len(pathParts) > 0 { // e.g. "/." or "/.."
				significantPart = pathParts[len(pathParts)-1]
			}

			if len(significantPart) > 0 {
				partToAdd = string(significantPart[0])
			}
		} else if !strings.HasPrefix(cleanArg, "-") { // Regular argument
			partToAdd = string(cleanArg[0])
		}

		if partToAdd != "" {
			if partToAdd == "." && len(nameParts) > 1 && len(strings.Join(nameParts, "")) > 1 {
				nameParts = append(nameParts, partToAdd) // Allow e.g. "ga."
			} else if partToAdd != "." {
				nameParts = append(nameParts, partToAdd)
			}
			if partToAdd != "." { // Don't count "." towards argInitialsCount.
				argInitialsCount++
			}
		}
	}

	if len(nameParts) == 1 { // Only command initial was added.
		if len(analyzedCmd.CommandName) >= 2 {
			return strings.ToLower(analyzedCmd.CommandName[:2]) // Use first two letters of command.
		}
		return nameParts[0] // Single letter command, single letter alias.
	}

	return strings.Join(nameParts, "")
}

func (g *AliasGenerator) generateExactFullCommandAliasesStrategy(
	commands []history.CommandFrequency,
	minFrequency int,
	existingAliases map[string]string,
	generatedNamesInThisRun map[string]bool, // Modifies this map
	minCommandEffectiveLength int,
) []alias.Alias {
	suggestions := []alias.Alias{}
	for _, cmdFreq := range commands {
		analyzed := g.analyzer.Analyze(cmdFreq.Command)

		if analyzed.IsComplex {
			// Optionally log: log.Printf("Skipping complex command for exact alias strategy: %s", cmdFreq.Command)
			continue
		}

		if analyzed.EffectiveLength < minCommandEffectiveLength {
			continue
		}

		if cmdFreq.Count < minFrequency {
			continue
		}

		if analyzed.CommandName == "" {
			continue
		}

		proposedName := g.generateExactCommandAliasName(analyzed)

		// Use the comprehensive IsValidAliasName for the final check,
		// after isProposedNameValid (which checks generatedInThisRun).
		// The current structure calls isProposedNameValid which is a preliminary check.
		// The full IsValidAliasName (with LookPath) is expected to be called by the service layer
		// or before finalizing. For internal generation, isProposedNameValid is used.
		if g.isProposedNameValid(proposedName, analyzed.CommandName, existingAliases, generatedNamesInThisRun) {
			suggestions = append(suggestions, alias.Alias{Name: proposedName, Command: cmdFreq.Command})
			generatedNamesInThisRun[proposedName] = true
		}
	}
	return suggestions
}

func (g *AliasGenerator) aggregateForCommandFirstArgStrategy(
	commands []history.CommandFrequency,
	minCommandEffectiveLength int,
) (map[string]int, map[string]command.AnalyzedCommand) {
	cmdFirstArgFreq := make(map[string]int)
	cmdFirstArgToAnalyzedCmd := make(map[string]command.AnalyzedCommand)

	for _, cmdFreq := range commands {
		analyzed := g.analyzer.Analyze(cmdFreq.Command)

		if analyzed.EffectiveLength < minCommandEffectiveLength {
			continue
		}

		if analyzed.CommandName == "" || len(analyzed.PotentialArgs) == 0 {
			continue
		}

		firstNonFlagArg := ""
		for _, arg := range analyzed.PotentialArgs {
			if !strings.HasPrefix(arg, "-") && len(arg) > 0 {
				firstNonFlagArg = arg
				break
			}
		}

		if firstNonFlagArg != "" {
			key := analyzed.CommandName + " " + firstNonFlagArg
			cmdFirstArgFreq[key] += cmdFreq.Count
			if _, exists := cmdFirstArgToAnalyzedCmd[key]; !exists {
				// Store a simplified AnalyzedCommand for generating the alias name.
				cmdFirstArgToAnalyzedCmd[key] = command.AnalyzedCommand{
					Original:        key, // The aggregated command string "cmd arg1"
					CommandName:     analyzed.CommandName,
					PotentialArgs:   []string{firstNonFlagArg},
					EffectiveLength: len(strings.ReplaceAll(key, " ", "")),
				}
			}
		}
	}
	return cmdFirstArgFreq, cmdFirstArgToAnalyzedCmd
}

func (g *AliasGenerator) generateAliasesFromCommandFirstArgAggregation(
	cmdFirstArgFreq map[string]int,
	cmdFirstArgToAnalyzedCmd map[string]command.AnalyzedCommand,
	minFrequency int,
	existingAliases map[string]string,
	generatedNamesInThisRun map[string]bool, // Modifies this map
) []alias.Alias {
	suggestions := []alias.Alias{}
	for keyCmdFirstArg, count := range cmdFirstArgFreq {
		if count < minFrequency {
			continue
		}

		analyzedForNameGen := cmdFirstArgToAnalyzedCmd[keyCmdFirstArg]
		proposedName := g.generateCommandSubcommandAliasName(analyzedForNameGen)
		aliasCommandString := keyCmdFirstArg // The alias command is the aggregated "cmd arg1"

		if g.isProposedNameValid(proposedName, analyzedForNameGen.CommandName, existingAliases, generatedNamesInThisRun) {
			suggestions = append(suggestions, alias.Alias{Name: proposedName, Command: aliasCommandString})
			generatedNamesInThisRun[proposedName] = true
		}
	}
	return suggestions
}
