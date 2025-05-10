package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
	"github.com/AntonioJCosta/nicksh/internal/handlers/ui"
	"github.com/spf13/cobra"
)

// NewAddPredefinedCommand creates the command for adding all predefined aliases.
func NewAddPredefinedCommand(
	suggestionSvc ports.AliasSuggestionService,
	managementSvc ports.AliasManagementService,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-predefined",
		Short: "Interactively adds predefined aliases from the configuration to your alias file.", // Modified
		Long: `Reads aliases from the predefined_aliases.yaml file,
validates them against your current shell aliases and system commands,
allows you to select which ones to add, and then adds them to your generated aliases file.`, // Modified
		RunE: func(cmd *cobra.Command, args []string) error {
			if suggestionSvc == nil {
				return fmt.Errorf("suggestion service is not initialized")
			}
			if managementSvc == nil {
				return fmt.Errorf("management service is not initialized")
			}

			currentShellAliases := loadCurrentShellAliases(managementSvc)

			validAliases, allLoadedAliases, err := fetchAndFilterPredefined(suggestionSvc, currentShellAliases)
			if err != nil {
				return err
			}

			if len(allLoadedAliases) == 0 {
				fmt.Println(ui.InfoColor("No predefined aliases found or loaded. Ensure 'predefined_aliases.yaml' exists and is readable, and the provider is configured."))
				return nil
			}

			if len(validAliases) == 0 {
				fmt.Println(ui.WarningColor(fmt.Sprintf("%d predefined aliases were found, but none were valid to add (they might conflict with existing aliases or system commands).", len(allLoadedAliases))))
				return nil
			}

			fmt.Println(ui.InfoColor(fmt.Sprintf("Found %d predefined aliases. %d are valid and available for selection:", len(allLoadedAliases), len(validAliases))))
			// Displaying aliases will be handled by fzf or numeric selection helpers

			var finalSelectedAliases []alias.Alias
			var selectionErr error

			// Attempt FZF selection first
			fzfSelected, fzfErr := selectAliasesViaFZF(validAliases) // from add_helpers.go

			if fzfErr == nil {
				finalSelectedAliases = fzfSelected
				if len(finalSelectedAliases) == 0 && len(validAliases) > 0 {
					fmt.Println(ui.InfoColor("No aliases selected via fzf."))
				}
			} else if errors.Is(fzfErr, ErrFZFNotFound) { // ErrFZFNotFound from add_helpers.go
				fmt.Println(ui.WarningColor("fzf not found in PATH. Falling back to numeric selection."))
				finalSelectedAliases, selectionErr = selectAliasesNumerically(validAliases) // from add_helpers.go
			} else if errors.Is(fzfErr, ErrFZFCancelled) { // ErrFZFCancelled from add_helpers.go
				fmt.Println(ui.InfoColor("Selection cancelled via fzf. No aliases will be added."))
				return nil // User cancelled
			} else {
				// Other fzf error, try numeric
				fmt.Fprintln(os.Stderr, ui.ErrorColor(fmt.Sprintf("Error during fzf selection: %v. Falling back to numeric selection.", fzfErr)))
				finalSelectedAliases, selectionErr = selectAliasesNumerically(validAliases) // from add_helpers.go
			}

			if selectionErr != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorColor(fmt.Sprintf("Error during alias selection: %v", selectionErr)))
				// Decide if you want to return or just print error and not add
				return nil
			}

			if len(finalSelectedAliases) == 0 {
				fmt.Println(ui.InfoColor("No aliases were selected to be added."))
				return nil
			}

			fmt.Println(ui.InfoColor(fmt.Sprintf("\nYou have selected %d predefined alias(es) to add.", len(finalSelectedAliases))))
			// Ask for confirmation for the selected aliases
			fmt.Print(ui.PromptColor(fmt.Sprintf("Do you want to add these %d selected aliases? (yes/no): ", len(finalSelectedAliases))))
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read user input: %w", err)
			}

			input = strings.TrimSpace(strings.ToLower(input))
			if input != "yes" && input != "y" {
				fmt.Println(ui.InfoColor("Aborted. No aliases were added."))
				return nil
			}

			fmt.Println(ui.InfoColor("Proceeding to add selected aliases..."))

			initiallyInvalidCount := len(allLoadedAliases) - len(validAliases) // This remains the same
			// Pass the user-selected aliases to addPredefinedToConfig
			successfullyAddedCount, skippedDueToExistingCount, addErrorCount := addPredefinedToConfig(finalSelectedAliases, managementSvc)

			// Adjust printAddPredefinedOutcome if its logic depends on "all valid" vs "selected"
			// For now, assuming it reports based on what was attempted to be added.
			printAddPredefinedOutcome(successfullyAddedCount, skippedDueToExistingCount, initiallyInvalidCount, addErrorCount, len(allLoadedAliases), managementSvc)

			return nil
		},
	}
	return cmd
}
