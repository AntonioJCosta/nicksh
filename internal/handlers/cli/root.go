package cli

import (
	"fmt"

	"github.com/AntonioJCosta/nicksh/internal/core/ports"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

func NewRootCommand(
	version string,
	suggestionService ports.AliasSuggestionService,
	managementService ports.AliasManagementService,
) *cobra.Command {
	rootCmd = &cobra.Command{
		Use:   "nicksh",
		Short: "nicksh helps you find and manage shell aliases.",
		Long: `nicksh analyzes your command history to suggest useful aliases
and provides tools to manage them in your shell configuration.`,
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if suggestionService == nil && (cmd.Name() == "suggest" || cmd.Name() == "add" || cmd.Name() == "add-predefined") {
				return fmt.Errorf("alias suggestion service not initialized for command %s", cmd.Name())
			}
			if managementService == nil && (cmd.Name() == "add" || cmd.Name() == "list" || cmd.Name() == "add-predefined") {
				return fmt.Errorf("alias management service not initialized for command %s", cmd.Name())
			}
			return nil
		},
	}

	rootCmd.AddCommand(NewSuggestCommand(suggestionService))
	rootCmd.AddCommand(NewAddCommand(suggestionService, managementService))
	rootCmd.AddCommand(NewListCommand(managementService))
	rootCmd.AddCommand(NewAddPredefinedCommand(suggestionService, managementService))

	return rootCmd
}
