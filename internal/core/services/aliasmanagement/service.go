package aliasmanagement

import (
	"fmt"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
)

type service struct {
	shellConfig ports.ShellConfigAccessor
}

// NewService creates a new alias management service.
// It panics if the shellConfigAccessor is nil.
func NewService(sc ports.ShellConfigAccessor) ports.AliasManagementService {
	if sc == nil {
		panic("shellConfig cannot be nil")
	}
	return &service{shellConfig: sc}
}

// AddAliasToConfig adds a new alias to the shell configuration.
// It returns true if the alias was newly added, false if it already existed (and was not overwritten),
// and an error if the operation failed.
func (s *service) AddAliasToConfig(name, command string) (bool, error) {
	if s.shellConfig == nil {
		// This check is defensive; NewService should prevent s.shellConfig from being nil.
		return false, fmt.Errorf("shellConfig is not initialized")
	}
	newAlias := alias.Alias{
		Name:    name,
		Command: command,
	}
	// Assuming s.shellConfig.AddAlias now returns (bool, error)
	// as per your internal/repositories/shellconfig/shell_config_accessor.go modification
	wasAdded, err := s.shellConfig.AddAlias(newAlias)
	if err != nil {
		return false, fmt.Errorf("failed to add alias '%s': %w", name, err)
	}
	return wasAdded, nil
}

// ListAliases retrieves all aliases currently managed by the shell configuration.
func (s *service) ListAliases() (map[string]string, error) {
	if s.shellConfig == nil {
		// Defensive check.
		return nil, fmt.Errorf("shellConfig is not initialized")
	}
	aliases, err := s.shellConfig.GetExistingAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to list existing aliases: %w", err)
	}
	return aliases, nil
}
