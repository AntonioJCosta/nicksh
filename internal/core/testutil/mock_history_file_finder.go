package testutil

import "github.com/AntonioJCosta/nicksh/internal/core/ports"

// MockHistoryFileFinder is a mock implementation of ports.HistoryFileFinder.
type MockHistoryFileFinder struct {
	FindFunc func() (string, error)
}

// Find mocks the Find method.
func (m *MockHistoryFileFinder) Find() (string, error) {
	if m.FindFunc != nil {
		return m.FindFunc()
	}
	return "", nil // Default behavior
}

var _ ports.HistoryFileFinder = (*MockHistoryFileFinder)(nil)
