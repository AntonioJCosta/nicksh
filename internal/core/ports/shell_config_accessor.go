package ports

import "github.com/AntonioJCosta/nicksh/internal/core/domain/alias"

/*
ShellConfigAccessor defines the interface for reading from and writing to
shell configuration files. This is a driven port, typically implemented by
a repository adapter that understands specific shell config formats.
*/
type ShellConfigAccessor interface {
	/*
	   GetExistingAliases retrieves all aliases currently defined in the relevant
	   shell configuration file(s).
	   It returns a map where the key is the alias name and the value is the command,
	   and an error if one occurred.
	*/
	GetExistingAliases() (map[string]string, error)

	/*
	   AddAlias appends a new alias to the appropriate shell configuration file.
	   newAlias is the Alias struct containing the name and command for the new alias.
	   It returns true if the alias was successfully added, false if it already exists,
	   and an error if one occurred.
	*/
	AddAlias(newAlias alias.Alias) (bool, error)
}
