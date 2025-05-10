package ports

import "github.com/AntonioJCosta/nicksh/internal/core/domain/alias"

// PredefinedAliasProvider defines the interface for sourcing aliases
// from a predefined list, like a configuration file.
type PredefinedAliasProvider interface {
	// GetPredefinedAliases loads aliases from a predefined source.
	GetPredefinedAliases() ([]alias.Alias, error)
}
