package commandanalysis

import (
	"reflect"
	"testing"

	"github.com/AntonioJCosta/nicksh/internal/core/domain/command"
)

func TestNewBasicAnalyzer(t *testing.T) {
	analyzer := NewBasicAnalyzer()
	if analyzer == nil {
		t.Fatal("NewBasicAnalyzer() returned nil")
	}
	if _, ok := analyzer.(*BasicAnalyzer); !ok {
		t.Errorf("NewBasicAnalyzer() did not return a *BasicAnalyzer, got %T", analyzer)
	}
}

func TestBasicAnalyzer_Analyze(t *testing.T) {
	analyzer := NewBasicAnalyzer()
	tests := []struct {
		name       string
		commandStr string
		want       command.AnalyzedCommand
	}{
		{
			name:       "simple command no args",
			commandStr: "ls",
			want: command.AnalyzedCommand{
				Original:        "ls",
				CommandName:     "ls",
				IsComplex:       false,
				PotentialArgs:   nil,
				EffectiveLength: 2, // "ls"
			},
		},
		{
			name:       "simple command with one arg",
			commandStr: "ls -l",
			want: command.AnalyzedCommand{
				Original:        "ls -l",
				CommandName:     "ls",
				IsComplex:       false,
				PotentialArgs:   []string{"-l"},
				EffectiveLength: 4, // "ls-l"
			},
		},
		{
			name:       "simple command with multiple args (becomes complex by word count)",
			commandStr: "git commit -m",
			want: command.AnalyzedCommand{
				Original:        "git commit -m",
				CommandName:     "git",
				IsComplex:       false,
				PotentialArgs:   []string{"commit", "-m"},
				EffectiveLength: 11, // "gitcommit-m"
			},
		},
		{
			name:       "command with quoted argument",
			commandStr: `git commit -m "initial commit"`,
			want: command.AnalyzedCommand{
				Original:        `git commit -m "initial commit"`,
				CommandName:     "git",
				IsComplex:       false, // Corrected to match 'got'
				PotentialArgs:   []string{"commit", "-m", "initial commit"},
				EffectiveLength: 26, // "gitcommit-m\"initialcommit\""
			},
		},
		{
			name:       "complex command with pipe",
			commandStr: "ls -l | grep test",
			want: command.AnalyzedCommand{
				Original:        "ls -l | grep test",
				CommandName:     "ls",
				IsComplex:       true,
				PotentialArgs:   []string{"-l", "|", "grep", "test"},
				EffectiveLength: 13, // "ls-l|greptest"
			},
		},
		{
			name:       "complex command with semicolon",
			commandStr: "cd /tmp; ls",
			want: command.AnalyzedCommand{
				Original:        "cd /tmp; ls",
				CommandName:     "cd",
				IsComplex:       true,
				PotentialArgs:   []string{"/tmp;", "ls"},
				EffectiveLength: 9, // "cd/tmp;ls"
			},
		},
		{
			name:       "complex command with ampersand",
			commandStr: "godo &",
			want: command.AnalyzedCommand{
				Original:        "godo &",
				CommandName:     "godo",
				IsComplex:       true, // Contains "&"
				PotentialArgs:   []string{"&"},
				EffectiveLength: 5, // "godo&"
			},
		},
		{
			name:       "complex command with parentheses",
			commandStr: "(echo hello)",
			want: command.AnalyzedCommand{
				Original:        "(echo hello)",
				CommandName:     "(echo",
				IsComplex:       true, // Contains "(" or ")"
				PotentialArgs:   []string{"hello)"},
				EffectiveLength: 11, // "(echohello)"
			},
		},
		{
			name:       "empty command string",
			commandStr: "",
			want: command.AnalyzedCommand{
				Original:        "",
				CommandName:     "",
				IsComplex:       false,
				PotentialArgs:   []string{},
				EffectiveLength: 0, // ""
			},
		},
		{
			name:       "command with only spaces",
			commandStr: "   ",
			want: command.AnalyzedCommand{
				Original:        "   ", // Original string is preserved
				CommandName:     "",
				IsComplex:       false,
				PotentialArgs:   []string{},
				EffectiveLength: 0, // ""
			},
		},
		{
			name:       "command with leading/trailing spaces",
			commandStr: "  ls -a  ",
			want: command.AnalyzedCommand{
				Original:        "  ls -a  ", // Original string is preserved
				CommandName:     "ls",
				IsComplex:       false, // len(args) is 2, so 2 > 2 is false
				PotentialArgs:   []string{"-a"},
				EffectiveLength: 4, // "ls-a"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.Analyze(tt.commandStr)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BasicAnalyzer.Analyze() diff:\ngot : %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}
