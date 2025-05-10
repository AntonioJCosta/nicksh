package cli

import (
	"fmt"
	"os"

	"github.com/AntonioJCosta/nicksh/internal/core/ports"
	"github.com/AntonioJCosta/nicksh/internal/handlers/ui"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// NewListCommand creates the 'list' subcommand.
func NewListCommand(aliasManagementService ports.AliasManagementService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List existing aliases managed by nicksh.",
		Long:  `Displays aliases found in the $HOME/.nicksh/ directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCmd(cmd, args, aliasManagementService)
		},
	}
	return cmd
}

// runListCmd contains the core logic for the 'list' command.
func runListCmd(
	_ *cobra.Command,
	_ []string,
	aliasManagementService ports.AliasManagementService,
) error {
	aliases, err := aliasManagementService.ListAliases()
	if err != nil {
		return fmt.Errorf("could not list aliases: %w", err)
	}

	if len(aliases) == 0 {
		fmt.Println(ui.InfoColor("No aliases found that are managed by nicksh in the $HOME/.nicksh/ directory."))
		return nil
	}

	fmt.Println(ui.HeaderColor("Existing Aliases (managed by nicksh in $HOME/.nicksh/):"))
	fmt.Println(ui.WarningColor("Note: These aliases are read from files in $HOME/.nicksh/."))
	fmt.Println(ui.WarningColor("       They reflect what nicksh manages, not necessarily your live shell's current alias state."))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Alias Name", "Command"})
	table.SetBorder(true)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})

	for name, command := range aliases {
		table.Append([]string{name, command})
	}
	table.Render()
	return nil
}
