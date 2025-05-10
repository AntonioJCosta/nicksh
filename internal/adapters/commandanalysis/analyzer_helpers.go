package commandanalysis

import (
	"strings"
	"unicode"
)

// parseArguments splits the command string into arguments,
// attempting to handle simple quoting and escape characters.
// This is a basic parser and may not cover all shell complexities.
func (a *BasicAnalyzer) parseArguments(trimmedCommandStr string) []string {
	var args []string
	var currentArg strings.Builder
	inQuotes := false
	isEscaped := false

	for _, r := range trimmedCommandStr {
		if isEscaped {
			currentArg.WriteRune(r) // Add the escaped character literally.
			isEscaped = false
			continue
		}

		switch r {
		case '\\':
			isEscaped = true
			// Backslash itself is not added, only its effect on the next character.
		case '"':
			inQuotes = !inQuotes
			// Quotes are delimiters and not part of the argument content.
		default:
			if unicode.IsSpace(r) && !inQuotes {
				if currentArg.Len() > 0 {
					args = append(args, currentArg.String())
					currentArg.Reset()
				}
			} else {
				currentArg.WriteRune(r)
			}
		}
	}
	if currentArg.Len() > 0 { // Add the last argument if any.
		args = append(args, currentArg.String())
	}
	return args
}

/*
determineComplexity provides an initial, simplified check for command complexity.

A command is currently considered complex if it:
 1. Consists of more than five parts (command + 4 arguments).
 2. Contains common shell metacharacters like |, &, ;, <, >, (, ).

This helps identify commands less suitable for simple, direct aliasing.
This definition may be refined in future versions.
*/
func (a *BasicAnalyzer) determineComplexity(originalCommandStr string, args []string) bool {
	// Assumes empty originalCommandStr is handled by the caller.
	numEffectiveParts := len(args)

	isComplexByArgCount := numEffectiveParts > 5
	containsShellChars := strings.ContainsAny(originalCommandStr, "|&;<>()")

	return isComplexByArgCount || containsShellChars
}
