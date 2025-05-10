package main

import (
	"fmt"
	"os"

	"github.com/AntonioJCosta/nicksh/internal/adapters/aliasgeneration"
	"github.com/AntonioJCosta/nicksh/internal/adapters/commandanalysis"
	"github.com/AntonioJCosta/nicksh/internal/adapters/oscommand"
	"github.com/AntonioJCosta/nicksh/internal/adapters/predefinedaliases"
	"github.com/AntonioJCosta/nicksh/internal/core/services/aliasmanagement"
	"github.com/AntonioJCosta/nicksh/internal/core/services/aliassuggestion"
	"github.com/AntonioJCosta/nicksh/internal/handlers/cli"
	"github.com/AntonioJCosta/nicksh/internal/repositories/history"
	"github.com/AntonioJCosta/nicksh/internal/repositories/shellconfig"
)

// Version is set at build time
var Version = "dev"

func main() {
	cmdExec := oscommand.NewOSCommandExecutor()

	historyFileFinder := history.NewDefaultHistoryFileFinder()
	historyRepo, err := history.NewHistoryProvider(cmdExec, historyFileFinder)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing history provider: %v\n", err)
		os.Exit(1)
	}

	cmdAnalyzer := commandanalysis.NewBasicAnalyzer()
	aliasGen := aliasgeneration.NewAliasGenerator(cmdAnalyzer)

	shellConf, err := shellconfig.NewShellConfigAccessor()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing shell config accessor: %v\n", err)
		os.Exit(1)
	}

	// predefinedAliasProvider can be nil if NewYAMLProvider returns an error
	predefinedAliasProvider, err := predefinedaliases.NewYAMLProvider()
	if err != nil {
		// The service will handle a nil predefinedAliasProvider.
		fmt.Fprintf(os.Stderr, "Warning: Could not initialize predefined alias provider %v. Continuing without predefined aliases.\n", err)
		predefinedAliasProvider = nil // Explicitly set to nil on error
	}
	// --- End Predefined Aliases Setup ---

	aliasSuggestionSvc := aliassuggestion.NewService(historyRepo, aliasGen, shellConf, predefinedAliasProvider) // Pass provider (can be nil)
	aliasManagementSvc := aliasmanagement.NewService(shellConf)
	rootCmd := cli.NewRootCommand(Version, aliasSuggestionSvc, aliasManagementSvc)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
