package testutil

import (
	"github.com/AntonioJCosta/nicksh/internal/core/domain/history"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
)

// MockHistoryProvider is a mock implementation of the ports.HistoryProvider interface.
type MockHistoryProvider struct {
	GetCommandFrequenciesFunc func(scanLimit int, outputLimit int) ([]history.CommandFrequency, error)
	GetHistoryFilePathFunc    func() string
	GetSourceIdentifierFunc   func() string
}

// GetCommandFrequencies mocks the GetCommandFrequencies method.
func (m *MockHistoryProvider) GetCommandFrequencies(scanLimit int, outputLimit int) ([]history.CommandFrequency, error) {
	if m.GetCommandFrequenciesFunc != nil {
		return m.GetCommandFrequenciesFunc(scanLimit, outputLimit)
	}
	// Default behavior: return empty slice and no error, or an error if that's more appropriate for your tests.
	return nil, nil
}

// GetHistoryFilePath mocks the GetHistoryFilePath method.
func (m *MockHistoryProvider) GetHistoryFilePath() string {
	if m.GetHistoryFilePathFunc != nil {
		return m.GetHistoryFilePathFunc()
	}
	// Default behavior: return an empty string.
	return ""
}

// GetSourceIdentifier mocks the GetSourceIdentifier method.
func (m *MockHistoryProvider) GetSourceIdentifier() string {
	if m.GetSourceIdentifierFunc != nil {
		return m.GetSourceIdentifierFunc()
	}
	// Default behavior: return an empty string.
	return ""
}

// Ensure MockHistoryProvider implements the ports.HistoryProvider interface.
var _ ports.HistoryProvider = (*MockHistoryProvider)(nil)
