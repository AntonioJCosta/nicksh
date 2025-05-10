package ports

import "github.com/AntonioJCosta/nicksh/internal/core/domain/command"

/*
CommandAnalyzer defines the contract for a service that analyzes a command string.
This is a driven port, representing a domain capability.
*/
type CommandAnalyzer interface {
	Analyze(commandStr string) command.AnalyzedCommand
}
