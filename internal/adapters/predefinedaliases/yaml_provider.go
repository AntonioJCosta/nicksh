package predefinedaliases

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
	"gopkg.in/yaml.v3"
)

//go:embed predefined_aliases.yaml
var embeddedPredefinedAliases []byte

// YAMLProvider implements the PredefinedAliasProvider interface
// by reading aliases from an embedded YAML file.
type YAMLProvider struct {
}

// NewYAMLProvider creates a new YAMLProvider.
// This implementation does not support local file overrides and always uses
// the embedded YAML file for predefined aliases. The filePath argument is not used.
// Future maintainers should update this function if local file overrides are needed.
func NewYAMLProvider() (ports.PredefinedAliasProvider, error) {
	// No filePath needed if always using embedded
	return &YAMLProvider{}, nil
}

// GetPredefinedAliases reads and parses aliases from the embedded YAML content.
func (p *YAMLProvider) GetPredefinedAliases() ([]alias.Alias, error) {
	predefined := []alias.Alias{}

	// Use the embedded content directly
	if len(embeddedPredefinedAliases) == 0 {
		// This case means the embedded file was empty or embedding failed (unlikely if file exists at compile time)
		return predefined, nil // No predefined aliases
	}

	decoder := yaml.NewDecoder(bytes.NewReader(embeddedPredefinedAliases))
	decoder.KnownFields(true) // Good practice to catch typos in the YAML structure

	err := decoder.Decode(&predefined)
	if err != nil {
		// Check if the error is EOF, which can happen with an empty file
		// or a file that only contains comments or is otherwise empty from YAML perspective.
		if errors.Is(err, io.EOF) {
			// Treat EOF as no aliases found.
			return []alias.Alias{}, nil // Return empty slice, not the potentially partially filled 'predefined'
		}
		return nil, fmt.Errorf("failed to unmarshal embedded predefined aliases: %w", err)
	}

	return predefined, nil
}
