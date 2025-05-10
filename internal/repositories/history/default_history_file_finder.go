package history

import "github.com/AntonioJCosta/nicksh/internal/core/ports"

// DefaultHistoryFileFinder is the default implementation that uses the package-level findUserHistoryFile.
type DefaultHistoryFileFinder struct{}

// Find implements the ports.HistoryFileFinder interface.
func (d *DefaultHistoryFileFinder) Find() (string, error) {
	return findUserHistoryFile() // Calls your existing global/package-level function
}

// NewDefaultHistoryFileFinder creates a new DefaultHistoryFileFinder.
func NewDefaultHistoryFileFinder() ports.HistoryFileFinder {
	return &DefaultHistoryFileFinder{}
}
