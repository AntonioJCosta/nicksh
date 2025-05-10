package ports

// If your AddAliasToConfig method in the service implementation
// or the underlying ShellConfigAccessor.AddAlias method uses
// the domain.Alias struct, you might need this import.
// For now, AddAliasToConfig takes simple strings, so it's not strictly needed here.
// import "github.com/AntonioJCosta/nicksh/internal/core/domain/alias"

// AliasManagementService defines the contract for managing shell aliases.
type AliasManagementService interface {
	// AddAliasToConfig adds a new alias to the shell configuration.
	// It returns true if the alias was newly added, false if it was skipped (e.g., already exists),
	// and an error if the operation failed.
	AddAliasToConfig(aliasName, aliasCommand string) (bool, error)

	// ListAliases retrieves all existing aliases from the shell configuration.
	ListAliases() (map[string]string, error)
}
