package aliasgeneration

import (
	"testing"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/command"
	"github.com/AntonioJCosta/nicksh/internal/core/testutil"
)

func TestNewAliasGenerator(t *testing.T) {
	mockAnalyzer := testutil.NewMockCommandAnalyzer()
	gen := NewAliasGenerator(mockAnalyzer)
	if gen == nil {
		t.Fatal("NewAliasGenerator returned nil")
	}
	if _, ok := gen.(*AliasGenerator); !ok {
		t.Fatalf("NewAliasGenerator did not return a *AliasGenerator, got %T", gen)
	}

}

func TestGenerateCommandSubcommandAliasName(t *testing.T) {
	gen := &AliasGenerator{} // This helper method does not depend on the analyzer field.

	tests := []struct {
		name        string
		analyzedCmd command.AnalyzedCommand
		want        string
	}{
		{
			name: "git status",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "git",
				PotentialArgs: []string{"status"},
			},
			want: "gs",
		},
		{
			name: "kubectl get",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "kubectl",
				PotentialArgs: []string{"get"},
			},
			want: "kg",
		},
		{
			name: "docker ps",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "docker",
				PotentialArgs: []string{"ps"},
			},
			want: "dp",
		},
		{
			name: "cd .. (special case)",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "cd",
				PotentialArgs: []string{".."},
			},
			want: "up",
		},
		{
			name: "command with path in first arg",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "run",
				PotentialArgs: []string{"./scripts/build.sh"},
			},
			want: "rb", // r from run, b from build.sh
		},
		{
			name: "command with path in first arg, no extension",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "execute",
				PotentialArgs: []string{"/usr/local/bin/mytool"},
			},
			want: "em", // e from execute, m from mytool
		},
		{
			name: "command with first arg being a flag",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "longcommand",
				PotentialArgs: []string{"-v", "otherarg"},
			},
			want: "lo", // Fallback to first two letters of command if first arg is a flag.
		},
		{
			name: "single char command with first arg being a flag",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "c",
				PotentialArgs: []string{"-v"},
			},
			want: "c", // Stays as 'c', length check is done by isProposedNameValid.
		},
		{
			name: "command name only, no args",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "mycommand",
				PotentialArgs: []string{},
			},
			want: "my",
		},
		{
			name: "short command name, no args",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "az",
				PotentialArgs: []string{},
			},
			want: "az",
		},
		{
			name: "command with one arg, mixed case",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "Go",
				PotentialArgs: []string{"Test"},
			},
			want: "gt",
		},
		{
			name:        "empty command name",
			analyzedCmd: command.AnalyzedCommand{CommandName: ""},
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gen.generateCommandSubcommandAliasName(tt.analyzedCmd); got != tt.want {
				t.Errorf("generateCommandSubcommandAliasName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateExactCommandAliasName(t *testing.T) {
	gen := &AliasGenerator{} // This helper method does not depend on the analyzer field.

	tests := []struct {
		name        string
		analyzedCmd command.AnalyzedCommand
		want        string
	}{
		{
			name: "git add .",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "git",
				PotentialArgs: []string{"add", "."},
			},
			want: "ga.",
		},
		{
			name: "kubectl get pods -n my-namespace",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "kubectl",
				PotentialArgs: []string{"get", "pods", "-n", "my-namespace"},
			},
			want: "kgpn", // k + g + p + n (from -n)
		},
		{
			name: "docker run --rm -it myimage",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "docker",
				PotentialArgs: []string{"run", "--rm", "-it", "myimage"},
			},
			want: "drri", // d + r + r (from --rm). Max 3 arg initials.
		},
		{
			name: "go test ./...",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "go",
				PotentialArgs: []string{"test", "./..."},
			},
			want: "gt.", // g + t + . (from ./...)
		},
		{
			name: "complex command falls back to first two letters of command name",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "verylongcommand",
				PotentialArgs: []string{"arg1", "arg2", "arg3"},
				IsComplex:     true,
			},
			want: "ve",
		},
		{
			name: "complex short command falls back to first letter",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "c",
				PotentialArgs: []string{"arg1"},
				IsComplex:     true,
			},
			want: "c", // Length check handled by isProposedNameValid.
		},
		{
			name: "single command, no args",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "mycommandlong",
				PotentialArgs: []string{},
			},
			want: "my",
		},
		{
			name: "command with only flags, max 2 arg initials",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "ls",
				PotentialArgs: []string{"-l", "-h", "-a"},
			},
			want: "llha", // ls + l (from -l) + a (from -a)
		},
		{
			name: "command with special char arg",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "grep",
				PotentialArgs: []string{"-R", "my_pattern*"},
			},
			want: "grm", // g + R + m
		},
		{
			name:        "empty command name",
			analyzedCmd: command.AnalyzedCommand{CommandName: ""},
			want:        "",
		},
		{
			name: "cd .. (special case for exact command)",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "cd",
				PotentialArgs: []string{".."},
			},
			want: "up", // "up" is the special case.
		},
		{
			name: "command with path arg, takes initial of last significant part",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "open",
				PotentialArgs: []string{"/some/long/path/to/file.txt"},
			},
			want: "of", // o + f (from file.txt)
		},
		{
			name: "command with path arg ending in /.",
			analyzedCmd: command.AnalyzedCommand{
				CommandName:   "explore",
				PotentialArgs: []string{"project/src/."},
			},
			want: "es", // e + s
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gen.generateExactCommandAliasName(tt.analyzedCmd); got != tt.want {
				t.Errorf("generateExactCommandAliasName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsProposedNameValid(t *testing.T) {
	gen := &AliasGenerator{} // This helper method does not depend on the analyzer field.
	existingAliases := map[string]string{
		"gp":  "git push",
		"old": "old command",
	}
	generatedInThisRun := map[string]bool{
		"ga": true,
	}

	tests := []struct {
		name                string
		proposedName        string
		originalCommandName string
		existingAliases     map[string]string
		generatedInThisRun  map[string]bool
		want                bool
	}{
		{"valid new name", "gl", "git", existingAliases, generatedInThisRun, true},
		{"invalid - too short", "g", "git", existingAliases, generatedInThisRun, false},
		{"invalid - contains invalid char", "g!", "git", existingAliases, generatedInThisRun, false},
		{"invalid - same as original command", "git", "git", existingAliases, generatedInThisRun, false},
		{"invalid - already generated in this run", "ga", "git", existingAliases, generatedInThisRun, false},
		{"invalid - conflicts with existing alias (proposed 'gp' for 'git' vs existing 'gp' for 'git push')", "gp", "git", existingAliases, generatedInThisRun, false},
		{"invalid - conflicts with existing alias (proposed 'gp' for 'gopher' vs existing 'gp' for 'git push')", "gp", "gopher", existingAliases, generatedInThisRun, false},
		{"valid - alphanumeric", "g1", "git", existingAliases, generatedInThisRun, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gen.isProposedNameValid(tt.proposedName, tt.originalCommandName, tt.existingAliases, tt.generatedInThisRun); got != tt.want {
				t.Errorf("isProposedNameValid() = %v, want %v for %s", got, tt.want, tt.name)
			}
		})
	}
}

// TestIsAliasNameValid (package-level helper, not method on AliasGenerator)
func TestIsAliasNameValid_PackageHelper(t *testing.T) {
	tests := []struct {
		name            string
		aliasName       string
		existingAliases map[string]string
		want            bool
	}{
		{
			name:            "valid name - not existing",
			aliasName:       "newalias",
			existingAliases: map[string]string{"oldalias": "cmd"},
			want:            true,
		},
		{
			name:            "invalid - existing alias",
			aliasName:       "myalias",
			existingAliases: map[string]string{"myalias": "some command"},
			want:            false,
		},
		{
			name:            "empty alias name - considered valid by this specific helper (length checked elsewhere)",
			aliasName:       "",
			existingAliases: map[string]string{},
			want:            true, // This helper only checks against the provided map.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAliasNameValid(tt.aliasName, tt.existingAliases); got != tt.want {
				t.Errorf("isAliasNameValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
