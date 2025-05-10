package predefinedaliases

import (
	"bytes"
	"errors" // Import errors package
	"fmt"
	"io" // Import io package for io.EOF
	"os"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
	"gopkg.in/yaml.v3"
)

// YAMLProvider implements the PredefinedAliasProvider interface
// by reading aliases from a YAML file.
type YAMLProvider struct {
	filePath string
}

// NewYAMLProvider creates a new YAMLProvider.
// filePath is the path to the YAML file containing predefined aliases.
func NewYAMLProvider(filePath string) (ports.PredefinedAliasProvider, error) {
	if filePath == "" {
		return nil, fmt.Errorf("YAML file path cannot be empty")
	}
	return &YAMLProvider{filePath: filePath}, nil
}

// GetPredefinedAliases reads and parses aliases from the configured YAML file.
// If the file does not exist or is empty, it returns an empty list and no error.
func (p *YAMLProvider) GetPredefinedAliases() ([]alias.Alias, error) {
	predefined := []alias.Alias{}

	yamlFile, err := os.ReadFile(p.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File not existing is not an error for this provider; it means no predefined aliases.
			return predefined, nil
		}
		return nil, fmt.Errorf("failed to read predefined aliases file %s: %w", p.filePath, err)
	}

	// If the file is empty, os.ReadFile returns an empty slice and no error.
	// An empty yamlFile will cause decoder.Decode to return io.EOF.
	if len(yamlFile) == 0 {
		return predefined, nil // Empty file means no aliases
	}

	decoder := yaml.NewDecoder(bytes.NewReader(yamlFile))
	decoder.KnownFields(true)

	err = decoder.Decode(&predefined)
	if err != nil {
		// Check if the error is EOF, which can happen with an empty file
		// or a file that only contains comments or is otherwise empty from YAML perspective.
		// We've already handled len(yamlFile) == 0, but a file with just "---" or comments
		// might also lead to EOF if no actual documents are found.
		if errors.Is(err, io.EOF) {
			// Treat EOF as no aliases found, similar to an empty or non-existent file.
			return predefined, nil
		}
		return nil, fmt.Errorf("failed to unmarshal predefined aliases from %s: %w", p.filePath, err)
	}

	return predefined, nil
}
