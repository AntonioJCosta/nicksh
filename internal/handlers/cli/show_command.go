package cli

import (
	"fmt"

	"github.com/AntonioJCosta/nicksh/internal/core/ports"
	"github.com/AntonioJCosta/nicksh/internal/handlers/ui"
	"github.com/spf13/cobra"
)

// NewSuggestCommand creates the 'show' subcommand.
func NewSuggestCommand(aliasSuggestionService ports.AliasSuggestionService) *cobra.Command {
	var minFrequency, scanLimit, outputLimit int

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show alias suggestions based on command history.",
		Long:  `Analyzes command history to find frequently used commands and suggests potential aliases.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowCmd(cmd, args, aliasSuggestionService)
		},
	}

	cmd.Flags().IntVarP(&minFrequency, "min-frequency", "f", 0, "Minimum frequency for a command to be considered for an alias (default 3).")
	cmd.Flags().IntVarP(&scanLimit, "scan-limit", "s", 0, "Number of recent history entries to scan (default 500).")
	cmd.Flags().IntVarP(&outputLimit, "output-limit", "o", 0, "Maximum number of alias suggestions to show (default 10).")

	return cmd
}

func runShowCmd(
	cmd *cobra.Command,
	_ []string,
	aliasSuggestionService ports.AliasSuggestionService,
) error {
	minFrequency, _ := cmd.Flags().GetInt("min-frequency")
	scanLimit, _ := cmd.Flags().GetInt("scan-limit")
	outputLimit, _ := cmd.Flags().GetInt("output-limit")

	// Default values
	if minFrequency <= 0 {
		minFrequency = 3
	}
	if scanLimit <= 0 {
		scanLimit = 500
	}
	if outputLimit <= 0 {
		outputLimit = 10
	}

	suggestionResult, err := aliasSuggestionService.GetSuggestions(minFrequency, scanLimit, outputLimit)
	if err != nil {
		return fmt.Errorf("could not get suggestions: %w", err)
	}

	if len(suggestionResult.Suggestions) == 0 {
		fmt.Println(ui.InfoColor("No alias suggestions found with the current criteria."))
		if suggestionResult.SourceDetails != "" {
			fmt.Println(ui.DetailColor(fmt.Sprintf("Context: %s", suggestionResult.SourceDetails)))
		} else {
			// Attempt to get context details if not already provided in the result.
			contextDetails, errCtx := aliasSuggestionService.GetSuggestionContextDetails()
			if errCtx == nil && contextDetails != "" {
				fmt.Println(ui.DetailColor(fmt.Sprintf("Context: %s", contextDetails)))
			}
		}
		return nil
	}

	fmt.Println(ui.InfoColor("Suggested Aliases:"))
	for _, s := range suggestionResult.Suggestions {
		fmt.Printf("  %s %s='%s'\n",
			ui.AliasKeywordColor("alias"),
			ui.AliasNameColor(s.Name),
			ui.AliasCmdColor(s.Command))
	}
	if suggestionResult.SourceDetails != "" {
		fmt.Println(ui.DetailColor(fmt.Sprintf("\n(Source: %s)", suggestionResult.SourceDetails)))
	}
	return nil
}
