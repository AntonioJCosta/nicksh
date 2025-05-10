package cli

import (
	"fmt"
	"os"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
	"github.com/AntonioJCosta/nicksh/internal/handlers/ui"
)

func loadCurrentShellAliases(managementSvc ports.AliasManagementService) map[string]string {
	fmt.Println(ui.InfoColor("Loading current shell aliases..."))
	currentShellAliases, err := managementSvc.ListAliases()
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.WarningColor(fmt.Sprintf("Warning: could not load current shell aliases: %v. Validation against them might be incomplete.", err)))
		return make(map[string]string)
	}
	return currentShellAliases
}

func fetchAndFilterPredefined(suggestionSvc ports.AliasSuggestionService, currentShellAliases map[string]string) ([]alias.Alias, []alias.Alias, error) {
	fmt.Println(ui.InfoColor("Fetching and filtering predefined aliases..."))
	validAliases, allLoadedAliases, err := suggestionSvc.GetFilteredPredefinedAliases(currentShellAliases)
	if err != nil {
		// This error will be returned and handled by the caller command.
		return nil, nil, fmt.Errorf("failed to get filtered predefined aliases: %w", err)
	}
	return validAliases, allLoadedAliases, nil
}

func addPredefinedToConfig(validAliases []alias.Alias, managementSvc ports.AliasManagementService) (successfullyAddedCount int, skippedDueToExistingCount int, addErrorCount int) {
	for _, pa := range validAliases {
		actuallyAdded, err := managementSvc.AddAliasToConfig(pa.Name, pa.Command)
		if err != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorColor(fmt.Sprintf("Error adding predefined alias '%s': %v", pa.Name, err)))
			addErrorCount++
		} else if actuallyAdded {
			successfullyAddedCount++
		} else {
			skippedDueToExistingCount++
		}
	}
	return successfullyAddedCount, skippedDueToExistingCount, addErrorCount
}

func printAddPredefinedOutcome(
	successfullyAddedCount int,
	skippedDueToExistingCount int,
	initiallyInvalidCount int,
	addErrorCount int,
	totalLoadedCount int,
	_ ports.AliasManagementService, // managementSvc is not used here, consider removing if not planned for future use.
) {
	if successfullyAddedCount > 0 {
		fmt.Println(ui.SuccessColor(fmt.Sprintf("\n%d predefined alias(es) successfully written to a file in the $HOME/.nicksh/ directory.", successfullyAddedCount)))

		if skippedDueToExistingCount > 0 {
			fmt.Println(ui.InfoColor(fmt.Sprintf("%d predefined alias(es) were skipped because they already exist.", skippedDueToExistingCount)))
		}
		if initiallyInvalidCount > 0 || addErrorCount > 0 {
			fmt.Println(ui.WarningColor(fmt.Sprintf("%d predefined alias(es) were skipped due to conflicts or failed to add.", initiallyInvalidCount+addErrorCount)))
		}

		// Instructions for the user.
		fmt.Println(ui.InfoColor("\nTo use the new alias(es):"))
		fmt.Println(ui.InfoColor("1. Ensure the following lines are present in your shell configuration file (e.g., ~/.bashrc, ~/.zshrc):"))
		fmt.Println(ui.CodeColor("   # Load all alias files from $HOME/.nicksh if the directory exists"))
		fmt.Println(ui.CodeColor(`   if [ -d "$HOME/.nicksh" ]; then`))
		fmt.Println(ui.CodeColor(`     for file in "$HOME/.nicksh"/*; do`))
		fmt.Println(ui.CodeColor(`       [ -f "$file" ] && source "$file"`))
		fmt.Println(ui.CodeColor("     done"))
		fmt.Println(ui.CodeColor("   fi"))
		fmt.Println(ui.InfoColor("\n2. Then, reload your shell configuration (e.g., 'source ~/.bashrc') or open a new terminal session."))

	} else {
		totalSkippedOrFailed := initiallyInvalidCount + addErrorCount + skippedDueToExistingCount
		if totalLoadedCount > 0 && totalSkippedOrFailed == totalLoadedCount {
			if skippedDueToExistingCount > 0 && initiallyInvalidCount == 0 && addErrorCount == 0 {
				fmt.Println(ui.InfoColor(fmt.Sprintf("\nNo new predefined aliases were added. All %d valid aliases already exist.", skippedDueToExistingCount)))
			} else {
				fmt.Println(ui.WarningColor(fmt.Sprintf("\nNo predefined aliases were added. All %d loaded aliases were skipped or failed to add.", totalLoadedCount)))
			}
		} else if totalLoadedCount == 0 {
			fmt.Println(ui.InfoColor("\nNo predefined aliases were found or loaded to process."))
		} else {
			// This case implies some were loaded, none added, but not all were skipped/failed.
			// Could happen if validAliases was empty but totalLoadedCount > 0.
			fmt.Println(ui.WarningColor("\nNo predefined aliases were successfully added."))
		}
	}
}
