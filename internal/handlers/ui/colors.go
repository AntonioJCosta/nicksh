package ui

import "github.com/fatih/color"

// General Purpose Colors
var (
	InfoColor    = color.New(color.FgCyan).SprintFunc()
	SuccessColor = color.New(color.FgGreen).SprintFunc()
	WarningColor = color.New(color.FgYellow).SprintFunc()
	ErrorColor   = color.New(color.FgRed).SprintFunc()
	PromptColor  = color.New(color.FgMagenta).SprintFunc()
	CodeColor    = color.New(color.FgWhite).SprintFunc()   // For code snippets
	DetailColor  = color.New(color.FgHiBlack).SprintFunc() // For less prominent details like source
)

// Alias Specific Colors
var (
	AliasKeywordColor = color.New(color.FgBlue, color.Bold).SprintFunc()
	AliasNameColor    = color.New(color.FgYellow).SprintFunc()
	AliasCmdColor     = color.New(color.FgWhite).SprintFunc()
)

// Header Colors
var (
	HeaderColor = color.New(color.FgGreen, color.Bold).SprintFunc()
)

// List Colors
var (
	ListItemColor = color.New(color.FgCyan).SprintFunc()
)
