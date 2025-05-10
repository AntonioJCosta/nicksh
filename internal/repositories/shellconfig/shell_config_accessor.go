package shellconfig

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
)

const generatedAliasesDir = ".nicksh"
const generatedAliasesFilename = "generated_aliases"

// userFriendlyGeneratedPath constructs a path string for display to the user.
func userFriendlyGeneratedPath() string {
	return filepath.Join("~/", generatedAliasesDir, generatedAliasesFilename)
}

// ShellConfigAccessor provides access to shell configuration files via the file system.
type ShellConfigAccessor struct {
	shell                    string
	generatedAliasesFilePath string
}

// NewShellConfigAccessor creates a new FileShellConfigAccessor.
func NewShellConfigAccessor() (ports.ShellConfigAccessor, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	homeDir := usr.HomeDir

	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return nil, fmt.Errorf("SHELL environment variable not set")
	}
	shellName := filepath.Base(shellPath)

	generatedAliasesDirFull := filepath.Join(homeDir, generatedAliasesDir)
	generatedAliasesFileFullPath := filepath.Join(generatedAliasesDirFull, generatedAliasesFilename)

	return &ShellConfigAccessor{
		shell:                    shellName,
		generatedAliasesFilePath: generatedAliasesFileFullPath,
	}, nil
}

// ...existing code...
// GetExistingAliases implements the ports.ShellConfigAccessor interface.
// It now reads all files from the $HOME/.nicksh/ directory.
func (sca *ShellConfigAccessor) GetExistingAliases() (map[string]string, error) {
	aliases := make(map[string]string)
	aliasesDir := filepath.Dir(sca.generatedAliasesFilePath) // Get the $HOME/.nicksh directory

	// Ensure the directory exists, but don't error if it doesn't; just return no aliases.
	if _, err := os.Stat(aliasesDir); os.IsNotExist(err) {
		return aliases, nil // No directory, so no aliases from it.
	}

	dirEntries, err := os.ReadDir(aliasesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read alias directory %s: %w", userFriendlyGeneratedPath(), err)
	}

	for _, entry := range dirEntries {
		if !entry.IsDir() {
			filePath := filepath.Join(aliasesDir, entry.Name())
			fileAliases, err := sca.getAliasesFromFile(filePath)
			if err != nil {
				// Log a warning but continue with other files
				fmt.Fprintf(os.Stderr, "Warning: could not read aliases from file %s: %v\n", toUserFriendlyPath(filePath), err)
				continue
			}
			for name, cmdVal := range fileAliases {
				if _, exists := aliases[name]; exists {
					// If an alias with the same name is found in multiple files,
					// log a warning. The last one read will take precedence.
					// Consider if a more sophisticated conflict resolution is needed.
					fmt.Fprintf(os.Stderr, "Warning: Alias '%s' found in multiple files. Using the definition from %s.\n", name, toUserFriendlyPath(filePath))
				}
				aliases[name] = cmdVal
			}
		}
	}

	return aliases, nil
}

// AddAlias implements the ports.ShellConfigAccessor interface.
// ...existing code...

// AddAlias implements the ports.ShellConfigAccessor interface.
func (sca *ShellConfigAccessor) AddAlias(newAlias alias.Alias) (bool, error) {
	dirPath := filepath.Dir(sca.generatedAliasesFilePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return false, fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	existingGeneratedAliases, err := sca.getAliasesFromFile(sca.generatedAliasesFilePath)
	if err != nil {
		return false, fmt.Errorf("failed to read existing generated aliases from %s: %w", userFriendlyGeneratedPath(), err)
	}

	if _, exists := existingGeneratedAliases[newAlias.Name]; exists {
		fmt.Printf("Alias '%s' already exists in %s. Skipping.\n", newAlias.Name, userFriendlyGeneratedPath())
		return false, nil
	}

	aliasLine := fmt.Sprintf("alias %s='%s'\n", newAlias.Name, newAlias.Command)

	file, err := os.OpenFile(sca.generatedAliasesFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false, fmt.Errorf("failed to open generated aliases file %s for appending: %w", userFriendlyGeneratedPath(), err)
	}
	defer file.Close()

	if _, err := file.WriteString(aliasLine); err != nil {
		return false, fmt.Errorf("failed to write alias to generated aliases file %s: %w", userFriendlyGeneratedPath(), err)
	}
	fmt.Printf("Alias '%s' added to %s.\n", newAlias.Name, userFriendlyGeneratedPath())
	return true, nil
}
