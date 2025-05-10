package testutil

import (
	"errors"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
)

// MockShellConfigAccessor is a mock implementation of ports.ShellConfigAccessor for testing.
type MockShellConfigAccessor struct {
	GetExistingAliasesFunc func() (map[string]string, error)
	AddAliasFunc           func(newAlias alias.Alias) (bool, error)
	GetConfigPathFunc      func() (string, error)
}

func (m *MockShellConfigAccessor) GetExistingAliases() (map[string]string, error) {
	if m.GetExistingAliasesFunc != nil {
		return m.GetExistingAliasesFunc()
	}
	return nil, errors.New("MockShellConfigAccessor: GetExistingAliasesFunc not implemented")
}

func (m *MockShellConfigAccessor) AddAlias(newAlias alias.Alias) (bool, error) {
	if m.AddAliasFunc != nil {
		return m.AddAliasFunc(newAlias)
	}
	return false, errors.New("MockShellConfigAccessor: AddAliasFunc not implemented")
}

func (m *MockShellConfigAccessor) GetConfigPath() (string, error) {
	if m.GetConfigPathFunc != nil {
		return m.GetConfigPathFunc()
	}
	return "", errors.New("MockShellConfigAccessor: GetConfigPathFunc not implemented")
}
