package commandanalysis

import (
	"strings"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/command"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
)

// BasicAnalyzer provides a simple implementation of command analysis.
type BasicAnalyzer struct{}

// NewBasicAnalyzer creates a new BasicAnalyzer.
func NewBasicAnalyzer() ports.CommandAnalyzer {
	return &BasicAnalyzer{}
}

// Analyze breaks down a command string into its components.
func (a *BasicAnalyzer) Analyze(commandStr string) command.AnalyzedCommand {
	trimmedCommandStr := strings.TrimSpace(commandStr)
	commandTextWithoutSpaces := strings.ReplaceAll(trimmedCommandStr, " ", "")
	effectiveLength := len(commandTextWithoutSpaces)

	if trimmedCommandStr == "" {
		return command.AnalyzedCommand{
			Original:        commandStr,
			CommandName:     "",
			IsComplex:       false, // An empty command is not complex.
			PotentialArgs:   []string{},
			EffectiveLength: 0,
		}
	}

	args := a.parseArguments(trimmedCommandStr)

	var cmdName string
	var potentialArgs []string

	if len(args) > 0 {
		cmdName = args[0]
		// Normalize command name by removing leading "./" if present.
		cmdName = strings.TrimPrefix(cmdName, "./")
		if len(args) > 1 {
			potentialArgs = args[1:]
		}
	}
	// else: args is empty, implies parseArguments had issues or input was unusual.
	// cmdName will be empty, which is handled by downstream logic.

	isComplex := a.determineComplexity(commandStr, args)

	return command.AnalyzedCommand{
		Original:        commandStr,
		CommandName:     cmdName,
		IsComplex:       isComplex,
		PotentialArgs:   potentialArgs,
		EffectiveLength: effectiveLength,
	}
}
