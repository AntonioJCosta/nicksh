package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
	"github.com/AntonioJCosta/nicksh/internal/handlers/ui"
	"github.com/spf13/cobra"
)

func NewAddCommand(
	aliasSuggestionService ports.AliasSuggestionService,
	aliasManagementService ports.AliasManagementService,
) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Interactively add suggested aliases to your shell configuration.",
		Long: `Shows alias suggestions and allows you to select which ones to add.
Uses fzf for selection if available, otherwise falls back to numeric input.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddCmd(cmd, args, aliasSuggestionService, aliasManagementService)
		},
	}

	cmd.Flags().IntP("min-frequency", "f", 0, "Minimum frequency for a command to be considered for an alias (default 3).")
	cmd.Flags().IntP("scan-limit", "s", 0, "Number of recent history entries to scan (default 500).")
	cmd.Flags().IntP("output-limit", "o", 0, "Maximum number of alias suggestions to show (default 10).")

	return cmd
}

func runAddCmd(
	cmd *cobra.Command,
	_ []string,
	aliasSuggestionService ports.AliasSuggestionService,
	aliasManagementService ports.AliasManagementService,
) error {
	flags := parseAddCommandFlags(cmd)

	if aliasSuggestionService == nil || aliasManagementService == nil {
		return fmt.Errorf("services not initialized for add command")
	}

	fmt.Println(ui.InfoColor("Fetching alias suggestions..."))
	suggestionResult, err := aliasSuggestionService.GetSuggestions(flags.minFrequency, flags.scanLimit, flags.outputLimit)
	if err != nil {
		return fmt.Errorf("could not get suggestions: %w", err)
	}

	if len(suggestionResult.Suggestions) == 0 {
		fmt.Println(ui.InfoColor("No alias suggestions found to add with the current criteria."))
		if suggestionResult.SourceDetails != "" {
			fmt.Println(ui.DetailColor(fmt.Sprintf("Context: %s", suggestionResult.SourceDetails)))
		}
		return nil
	}
	fmt.Println(ui.InfoColor(fmt.Sprintf("Found %d suggestions. (Source: %s)", len(suggestionResult.Suggestions), ui.DetailColor(suggestionResult.SourceDetails))))

	var finalSelectedAliases []alias.Alias
	var selectionErr error

	fzfSelected, fzfErr := selectAliasesViaFZF(suggestionResult.Suggestions)

	if fzfErr == nil {
		finalSelectedAliases = fzfSelected
		if len(finalSelectedAliases) == 0 && len(suggestionResult.Suggestions) > 0 {
			fmt.Println(ui.InfoColor("No aliases selected via fzf."))
		}
	} else if errors.Is(fzfErr, ErrFZFNotFound) {
		fmt.Println(ui.WarningColor("fzf not found in PATH. Falling back to numeric selection."))
		finalSelectedAliases, selectionErr = selectAliasesNumerically(suggestionResult.Suggestions)
	} else if errors.Is(fzfErr, ErrFZFCancelled) {
		fmt.Println(ui.InfoColor("Selection cancelled via fzf. No aliases will be added."))
	} else {
		fmt.Fprintln(os.Stderr, ui.ErrorColor(fmt.Sprintf("Error during fzf selection: %v. Falling back to numeric selection.", fzfErr)))
		finalSelectedAliases, selectionErr = selectAliasesNumerically(suggestionResult.Suggestions)
	}

	if selectionErr != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorColor(fmt.Sprintf("Error during alias selection: %v", selectionErr)))
		return nil
	}

	if len(finalSelectedAliases) == 0 {
		return nil
	}

	fmt.Println(ui.InfoColor(fmt.Sprintf("\nYou have selected %d alias(es) to add.", len(finalSelectedAliases))))

	successfullyAddedCount, skippedDueToExistingCount, addOutcomeErr := addAliasesToConfigAndPrintOutcome(finalSelectedAliases, aliasManagementService)

	if addOutcomeErr != nil {
		return fmt.Errorf("encountered an error while processing aliases (added: %d, skipped: %d): %w", successfullyAddedCount, skippedDueToExistingCount, addOutcomeErr)
	}

	if successfullyAddedCount == 0 && skippedDueToExistingCount == 0 && len(finalSelectedAliases) > 0 {
		fmt.Println(ui.WarningColor("\nNo aliases were successfully added or skipped from your selection (check for errors printed above)."))
	}

	return nil
}
