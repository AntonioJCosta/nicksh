package cli

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/alias"
	"github.com/AntonioJCosta/nicksh/internal/core/ports"
	"github.com/AntonioJCosta/nicksh/internal/handlers/ui"
	"github.com/spf13/cobra"
)

// ErrFZFNotFound indicates that the fzf binary was not found in PATH.
var ErrFZFNotFound = errors.New("fzf binary not found in PATH")

// ErrFZFCancelled indicates that the user cancelled the fzf selection (e.g., by pressing Esc or Ctrl-C).
var ErrFZFCancelled = errors.New("fzf selection cancelled by user")

type addCommandFlags struct {
	minFrequency int
	scanLimit    int
	outputLimit  int
}

func parseAddCommandFlags(cmd *cobra.Command) addCommandFlags {
	minFreq, _ := cmd.Flags().GetInt("min-frequency")
	scanLim, _ := cmd.Flags().GetInt("scan-limit")
	outLim, _ := cmd.Flags().GetInt("output-limit")

	// Default values if not provided or zero
	if minFreq == 0 {
		minFreq = 3
	}
	if scanLim == 0 {
		scanLim = 500
	}
	if outLim == 0 {
		outLim = 10
	}

	return addCommandFlags{
		minFrequency: minFreq,
		scanLimit:    scanLim,
		outputLimit:  outLim,
	}
}

func selectAliasesViaFZF(suggestions []alias.Alias) ([]alias.Alias, error) {
	fzfPath, err := exec.LookPath("fzf")
	if err != nil {
		return nil, ErrFZFNotFound
	}

	if len(suggestions) == 0 {
		return []alias.Alias{}, nil
	}

	var inputBuffer bytes.Buffer
	suggestionMap := make(map[string]alias.Alias)

	for _, s := range suggestions {
		// Feed raw alias strings to fzf for reliable mapping of selections.
		rawLine := fmt.Sprintf("alias %s='%s'", s.Name, s.Command)
		suggestionMap[rawLine] = s
		inputBuffer.WriteString(rawLine + "\n")
	}

	// The --ansi flag allows fzf to render ANSI codes if present in the prompt or preview.
	fzfCmd := exec.Command(fzfPath, "--multi", "--ansi", "--prompt", ui.PromptColor("Select aliases (TAB to multi-select, Enter to confirm) > "))
	fzfCmd.Stdin = &inputBuffer

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer
	fzfCmd.Stdout = &outBuffer
	fzfCmd.Stderr = &errBuffer

	err = fzfCmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 130 indicates user cancellation (e.g., Ctrl-C, Esc).
			if exitErr.ExitCode() == 130 {
				return nil, ErrFZFCancelled
			}
			// Exit code 1 with no output often means no selection was made.
			if exitErr.ExitCode() == 1 && strings.TrimSpace(outBuffer.String()) == "" {
				return []alias.Alias{}, nil
			}
		}
		return nil, fmt.Errorf("fzf execution failed (stderr: %s): %w", strings.TrimSpace(errBuffer.String()), err)
	}

	selectedLinesStr := strings.TrimSpace(outBuffer.String())
	if selectedLinesStr == "" {
		return []alias.Alias{}, nil
	}

	selectedLines := strings.Split(selectedLinesStr, "\n")
	var chosenAliases []alias.Alias
	for _, line := range selectedLines {
		trimmedLine := strings.TrimSpace(line)
		if selectedAlias, ok := suggestionMap[trimmedLine]; ok { // Map uses raw lines
			chosenAliases = append(chosenAliases, selectedAlias)
		} else if trimmedLine != "" {
			// This might occur if fzf's output format changes unexpectedly.
			fmt.Fprintln(os.Stderr, ui.WarningColor(fmt.Sprintf("Warning: fzf selected an unknown line: %s", trimmedLine)))
		}
	}

	return chosenAliases, nil
}

func displaySuggestionsForNumericSelection(suggestions []alias.Alias) {
	fmt.Println(ui.PromptColor("Select aliases to add (e.g., 1,3-5, or 'all', 'none'):"))
	for i, s := range suggestions {
		fmt.Printf("%d. %s %s='%s'\n",
			i+1,
			ui.AliasKeywordColor("alias"),
			ui.AliasNameColor(s.Name),
			ui.AliasCmdColor(s.Command))
	}
}

func parseNumericSelectionInput(input string, suggestionCount int) ([]int, error) {
	trimmedInput := strings.TrimSpace(strings.ToLower(input))
	if trimmedInput == "none" {
		return []int{}, nil
	}
	if trimmedInput == "all" {
		if suggestionCount == 0 {
			return []int{}, nil
		}
		indices := make([]int, suggestionCount)
		for i := 0; i < suggestionCount; i++ {
			indices[i] = i
		}
		return indices, nil
	}

	var selections []int
	parts := strings.Split(trimmedInput, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeParts := strings.SplitN(part, "-", 2)
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}
			start, err1 := strconv.Atoi(rangeParts[0])
			end, err2 := strconv.Atoi(rangeParts[1])
			if err1 != nil || err2 != nil || start <= 0 || end < start || end > suggestionCount {
				return nil, fmt.Errorf("invalid range or number (max %d): %s", suggestionCount, part)
			}
			for i := start; i <= end; i++ {
				selections = append(selections, i-1) // Convert to 0-based index
			}
		} else {
			num, err := strconv.Atoi(part)
			if err != nil || num <= 0 || num > suggestionCount {
				return nil, fmt.Errorf("invalid number (max %d): %s", suggestionCount, part)
			}
			selections = append(selections, num-1) // Convert to 0-based index
		}
	}

	// Ensure unique selections
	uniqueSelectionsMap := make(map[int]bool)
	var uniqueSelectionIndices []int
	for _, idx := range selections {
		if !uniqueSelectionsMap[idx] {
			uniqueSelectionsMap[idx] = true
			uniqueSelectionIndices = append(uniqueSelectionIndices, idx)
		}
	}
	return uniqueSelectionIndices, nil
}

func selectAliasesNumerically(suggestions []alias.Alias) ([]alias.Alias, error) {
	if len(suggestions) == 0 {
		return []alias.Alias{}, nil
	}

	displaySuggestionsForNumericSelection(suggestions)
	fmt.Print(ui.PromptColor("Your choice: "))
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read selection: %w", err)
	}

	selectedIndices, err := parseNumericSelectionInput(input, len(suggestions))
	if err != nil {
		return nil, fmt.Errorf("invalid selection input: %w", err)
	}

	var chosenAliases []alias.Alias
	for _, idx := range selectedIndices {
		if idx >= 0 && idx < len(suggestions) {
			chosenAliases = append(chosenAliases, suggestions[idx])
		}
	}
	return chosenAliases, nil
}

func addAliasesToConfigAndPrintOutcome(
	selectedAliases []alias.Alias,
	aliasManagementService ports.AliasManagementService,
) (successfullyAddedCount int, skippedDueToExistingCount int, firstError error) {

	if len(selectedAliases) == 0 {
		return 0, 0, nil
	}

	fmt.Println(ui.InfoColor("\nProcessing selected aliases..."))
	for _, selectedAlias := range selectedAliases {
		wasAdded, err := aliasManagementService.AddAliasToConfig(selectedAlias.Name, selectedAlias.Command)
		if err != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorColor(fmt.Sprintf("Error processing alias '%s': %v", selectedAlias.Name, err)))
			if firstError == nil {
				firstError = err
			}
		} else {
			if wasAdded {
				successfullyAddedCount++
			} else {
				skippedDueToExistingCount++
			}
		}
	}

	if successfullyAddedCount > 0 {
		fmt.Println(ui.SuccessColor(fmt.Sprintf("\n%d alias(es) successfully written to a file in the $HOME/.nicksh/ directory.", successfullyAddedCount)))
		if skippedDueToExistingCount > 0 {
			fmt.Println(ui.InfoColor(fmt.Sprintf("%d alias(es) were skipped because they already exist.", skippedDueToExistingCount)))
		}

		// Instructions for the user to source the new aliases.
		fmt.Println(ui.InfoColor("\nTo use the new alias(es):"))
		fmt.Println(ui.InfoColor("1. Ensure the following lines are present in your shell configuration file (e.g., ~/.bashrc, ~/.zshrc):"))
		fmt.Println(ui.CodeColor("   # Load all alias files from $HOME/.nicksh if the directory exists"))
		fmt.Println(ui.CodeColor(`   if [ -d "$HOME/.nicksh" ]; then`))
		fmt.Println(ui.CodeColor(`     for file in "$HOME/.nicksh"/*; do`))
		fmt.Println(ui.CodeColor(`       [ -f "$file" ] && source "$file"`))
		fmt.Println(ui.CodeColor("     done"))
		fmt.Println(ui.CodeColor("   fi"))
		fmt.Println(ui.InfoColor("\n2. Then, reload your shell configuration (e.g., 'source ~/.bashrc') or open a new terminal session."))

	} else if skippedDueToExistingCount > 0 && firstError == nil {
		fmt.Println(ui.InfoColor(fmt.Sprintf("\nNo new aliases were added. %d alias(es) from your selection already exist.", skippedDueToExistingCount)))
	} else if firstError != nil {
		// Specific errors were printed in the loop above.
		fmt.Println(ui.WarningColor("\nNo aliases were successfully added due to errors (see details above)."))
		if skippedDueToExistingCount > 0 {
			fmt.Println(ui.InfoColor(fmt.Sprintf("%d alias(es) were also skipped because they already exist.", skippedDueToExistingCount)))
		}
	}
	return successfullyAddedCount, skippedDueToExistingCount, firstError
}
