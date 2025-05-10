package testutil

import (
	"github.com/AntonioJCosta/nicksh/internal/core/domain/command"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
)

// MockCommandAnalyzer is a mock implementation of ports.CommandAnalyzer.
type MockCommandAnalyzer struct {
	// AnalyzeFunc allows you to set a custom function for the Analyze method.
	AnalyzeFunc func(commandStr string) command.AnalyzedCommand
	// AnalyzeCalls keeps track of the arguments passed to Analyze.
	AnalyzeCalls []string
}

// NewMockCommandAnalyzer creates a new MockCommandAnalyzer.
func NewMockCommandAnalyzer() *MockCommandAnalyzer {
	return &MockCommandAnalyzer{
		AnalyzeCalls: make([]string, 0),
	}
}

// Analyze implements the ports.CommandAnalyzer interface.
// It calls AnalyzeFunc if it's set, otherwise returns a zero-value AnalyzedCommand.
func (m *MockCommandAnalyzer) Analyze(commandStr string) command.AnalyzedCommand {
	m.AnalyzeCalls = append(m.AnalyzeCalls, commandStr)
	if m.AnalyzeFunc != nil {
		return m.AnalyzeFunc(commandStr)
	}
	// Default behavior if AnalyzeFunc is not set:
	// Return an empty/zero AnalyzedCommand or panic, depending on desired strictness.
	// For tests, returning a zero value is often fine if the test explicitly sets AnalyzeFunc.
	return command.AnalyzedCommand{}
}

// Ensure MockCommandAnalyzer satisfies the CommandAnalyzer interface.
var _ ports.CommandAnalyzer = (*MockCommandAnalyzer)(nil)
