package shellconfig

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func (sca *ShellConfigAccessor) getAliasesFromFile(filePath string) (map[string]string, error) {
	aliases := make(map[string]string)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return aliases, nil // File not existing is not an error for reading, just means no aliases there yet
		}
		return nil, fmt.Errorf("failed to open alias file %s: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		name, command, isAlias := parseAliasLineFromString(scanner.Text())
		if isAlias {
			aliases[name] = command
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning alias file %s: %w", filePath, err)
	}
	return aliases, nil
}

// parseAliasLineFromString remains an internal helper

// ...existing code...
func parseAliasLineFromString(line string) (name string, command string, isAlias bool) {
	trimmedLine := strings.TrimSpace(line)

	if strings.HasPrefix(trimmedLine, "#") {
		return "", "", false // It's a comment
	}

	if !strings.HasPrefix(trimmedLine, "alias ") {
		return "", "", false // Not an alias definition
	}

	// Remove "alias " prefix
	content := strings.TrimPrefix(trimmedLine, "alias ")

	// Split into name and value by the first '='
	parts := strings.SplitN(content, "=", 2)
	if len(parts) < 2 {
		// No '=' found after "alias name", so it's not a complete alias definition like "alias name=command"
		// Example: "alias foo" is not processed here as a full alias with a command.
		return "", "", false
	}

	name = strings.TrimSpace(parts[0])
	commandValue := strings.TrimSpace(parts[1])

	// Handle quoted command values
	if len(commandValue) >= 2 {
		firstChar := commandValue[0]
		lastChar := commandValue[len(commandValue)-1]

		if (firstChar == '\'' && lastChar == '\'') || (firstChar == '"' && lastChar == '"') {
			command = commandValue[1 : len(commandValue)-1]
		} else {
			command = commandValue // Not enclosed in matching quotes, or quotes are internal
		}
	} else {
		command = commandValue // Command is empty or a single character (cannot be quoted)
	}

	// According to tests:
	// "alias myls=" -> name="myls", command="", isAlias=true
	// "alias ="ls -l"" -> name="", command="ls -l", isAlias=true
	// So, an empty name OR an empty command is permissible if the structure is `alias name=command`.
	// The SplitN and subsequent trims handle this.
	// The critical part is that `SplitN` found an "=".

	// If name is empty, but we have a command part (e.g. "alias =foo"), it's considered an alias.
	// If name is present, but command part is empty (e.g. "alias foo="), it's considered an alias.
	// If both name and part[0] of command are empty after trim (e.g. "alias ="), it's an alias.

	return name, command, true
}

// ...existing code...

// Helper function (if not already present or imported)
func toUserFriendlyPath(absPath string) string {
	usr, err := user.Current()
	if err != nil {
		return absPath
	}
	homeDir := usr.HomeDir
	if strings.HasPrefix(absPath, homeDir) {
		if absPath == homeDir {
			return "~"
		}
		return filepath.Join("~", strings.TrimPrefix(absPath, homeDir+string(os.PathSeparator)))
	}
	return absPath
}
