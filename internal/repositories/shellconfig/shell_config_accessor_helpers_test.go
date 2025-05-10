package shellconfig

import (
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// manageTestFile creates a file at the given path for the test and ensures it's cleaned up.
// If content is empty, an empty file is created.
func manageTestFile(t *testing.T, path string, content []byte) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("Failed to create test file %s: %v", path, err)
	}
	t.Cleanup(func() {
		if err := os.Remove(path); err != nil {
			t.Logf("Warning: failed to remove test file %s: %v", path, err)
		}
	})
}

func TestParseAliasLineFromString(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantName    string
		wantCommand string
		wantIsAlias bool
	}{
		{
			name:        "valid alias with double quotes",
			line:        `alias ll="ls -alF"`,
			wantName:    "ll",
			wantCommand: "ls -alF",
			wantIsAlias: true,
		},
		{
			name:        "valid alias with single quotes",
			line:        `alias gp='git push'`,
			wantName:    "gp",
			wantCommand: "git push",
			wantIsAlias: true,
		},
		{
			name:        "valid alias with no quotes around command",
			line:        `alias g=git`,
			wantName:    "g",
			wantCommand: "git",
			wantIsAlias: true,
		},
		{
			name:        "valid alias with spaces around equals",
			line:        `alias   ga   =  "git add"`,
			wantName:    "ga",
			wantCommand: "git add",
			wantIsAlias: true,
		},
		{
			name:        "valid alias with leading/trailing spaces on line",
			line:        `  alias k=kubectl  `,
			wantName:    "k",
			wantCommand: "kubectl",
			wantIsAlias: true,
		},
		{
			name:        "alias with empty command double quotes",
			line:        `alias e=""`,
			wantName:    "e",
			wantCommand: "",
			wantIsAlias: true,
		},
		{
			name:        "alias with empty command single quotes",
			line:        `alias es=''`,
			wantName:    "es",
			wantCommand: "",
			wantIsAlias: true,
		},
		{
			name:        "comment line",
			line:        `# alias l="ls -CF"`,
			wantName:    "",
			wantCommand: "",
			wantIsAlias: false,
		},
		{
			name:        "empty line",
			line:        ``,
			wantName:    "",
			wantCommand: "",
			wantIsAlias: false,
		},
		{
			name:        "whitespace line",
			line:        `   `,
			wantName:    "",
			wantCommand: "",
			wantIsAlias: false,
		},
		{
			name:        "not an alias line",
			line:        `export PATH="/usr/local/bin:$PATH"`,
			wantName:    "",
			wantCommand: "",
			wantIsAlias: false,
		},
		{
			name:        "malformed alias - no equals",
			line:        `alias myls`,
			wantName:    "",
			wantCommand: "",
			wantIsAlias: false,
		},
		{
			name:        "malformed alias - no command",
			line:        `alias myls=`,
			wantName:    "myls",
			wantCommand: "",
			wantIsAlias: true,
		},
		{
			name:        "malformed alias - no name",
			line:        `alias ="ls -l"`,
			wantName:    "",
			wantCommand: "ls -l",
			wantIsAlias: true, // Current parser allows empty name if format is `alias =cmd`
		},
		{
			name:        "alias with complex command and internal quotes",
			line:        `alias glog="git log --graph --pretty=format:'%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%cr) %C(bold blue)<%an>%Creset' --abbrev-commit"`,
			wantName:    "glog",
			wantCommand: "git log --graph --pretty=format:'%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%cr) %C(bold blue)<%an>%Creset' --abbrev-commit",
			wantIsAlias: true,
		},
		{
			name:        "alias with single quote inside double quoted command",
			line:        `alias test="echo 'hello'"`,
			wantName:    "test",
			wantCommand: "echo 'hello'",
			wantIsAlias: true,
		},
		{
			name:        "alias with double quote inside single quoted command",
			line:        `alias test='echo "hello"'`,
			wantName:    "test",
			wantCommand: `echo "hello"`,
			wantIsAlias: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotCommand, gotIsAlias := parseAliasLineFromString(tt.line)
			if gotName != tt.wantName {
				t.Errorf("parseAliasLineFromString() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotCommand != tt.wantCommand {
				t.Errorf("parseAliasLineFromString() gotCommand = %v, want %v", gotCommand, tt.wantCommand)
			}
			if gotIsAlias != tt.wantIsAlias {
				t.Errorf("parseAliasLineFromString() gotIsAlias = %v, want %v", gotIsAlias, tt.wantIsAlias)
			}
		})
	}
}

func TestGetAliasesFromFile(t *testing.T) {
	sca := &ShellConfigAccessor{} // The method doesn't use sca's fields, so a simple instance is fine.
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		filePath    func() string // Function to generate path, for non-existent cases
		fileContent []byte
		wantAliases map[string]string
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name: "file does not exist",
			filePath: func() string {
				return filepath.Join(tempDir, "non_existent_aliases.txt")
			},
			fileContent: nil,
			wantAliases: map[string]string{},
			wantErr:     false,
		},
		{
			name: "empty file",
			filePath: func() string {
				path := filepath.Join(tempDir, "empty_aliases.txt")
				manageTestFile(t, path, []byte{})
				return path
			},
			fileContent: []byte{}, // Content managed by filePath func
			wantAliases: map[string]string{},
			wantErr:     false,
		},
		{
			name: "file with valid aliases",
			filePath: func() string {
				path := filepath.Join(tempDir, "valid_aliases.txt")
				content := `
alias ls='ls -G'
alias ll="ls -alF"
alias ..="cd .."
`
				manageTestFile(t, path, []byte(content))
				return path
			},
			wantAliases: map[string]string{
				"ls": "ls -G",
				"ll": "ls -alF",
				"..": "cd ..",
			},
			wantErr: false,
		},
		{
			name: "file with mixed content (aliases, comments, empty lines)",
			filePath: func() string {
				path := filepath.Join(tempDir, "mixed_aliases.txt")
				content := `
# This is a comment
alias g=git
alias ga="git add"

export SOME_VAR="value" # Not an alias
alias gl='git log --oneline'
`
				manageTestFile(t, path, []byte(content))
				return path
			},
			wantAliases: map[string]string{
				"g":  "git",
				"ga": "git add",
				"gl": "git log --oneline",
			},
			wantErr: false,
		},
		{
			name: "file with malformed aliases",
			filePath: func() string {
				path := filepath.Join(tempDir, "malformed_aliases.txt")
				content := `
alias ok1="command1"
alias noequals
alias ok2='command2'
alias = "missingname"
alias emptyname=
`
				manageTestFile(t, path, []byte(content))
				return path
			},
			wantAliases: map[string]string{
				"ok1":       "command1",
				"ok2":       "command2",
				"":          "missingname", // Current parser allows empty name
				"emptyname": "",
			},
			wantErr: false,
		},
		// Note: Testing os.Open failure for reasons other than IsNotExist (e.g. permissions)
		// is harder to do reliably in a cross-platform way without more complex test setup.
		// The current test covers the IsNotExist path.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.filePath() // Generate/setup file path

			gotAliases, err := sca.getAliasesFromFile(filePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("getAliasesFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("getAliasesFromFile() error msg = %q, want to contain %q", err.Error(), tt.wantErrMsg)
				}
				return // No need to check aliases if an error was expected and occurred
			}

			if !reflect.DeepEqual(gotAliases, tt.wantAliases) {
				t.Errorf("getAliasesFromFile() gotAliases = %v, want %v", gotAliases, tt.wantAliases)
			}
		})
	}
}

func TestToUserFriendlyPath(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user for testing: %v", err)
	}
	homeDir := currentUser.HomeDir

	if homeDir == "" {
		t.Log("Warning: Home directory is empty, toUserFriendlyPath might behave unexpectedly.")
	}

	tests := []struct {
		name    string
		absPath string
		want    string
	}{
		{
			name:    "path is exactly home directory",
			absPath: homeDir,
			want:    "~",
		},
		{
			name:    "path is a subdirectory of home",
			absPath: filepath.Join(homeDir, "Documents", "file.txt"),
			want:    filepath.Join("~", "Documents", "file.txt"),
		},
		{
			name:    "path is a file directly in home",
			absPath: filepath.Join(homeDir, ".bashrc"),
			want:    filepath.Join("~", ".bashrc"),
		},
		{
			name:    "path is outside home directory",
			absPath: "/usr/local/bin/someapp",
			want:    "/usr/local/bin/someapp",
		},
		{
			name:    "path is root directory",
			absPath: "/",
			want:    "/",
		},
		{
			name:    "path is empty",
			absPath: "",
			want:    "",
		},
		{
			name:    "path is a different user's home",
			absPath: "/home/otheruser/docs",
			want:    "/home/otheruser/docs",
		},
		{
			name:    "path with trailing slash in home",
			absPath: filepath.Join(homeDir, "subdir") + string(filepath.Separator),
			want:    filepath.Join("~", "subdir"),
		},
		{
			name:    "path is home directory with trailing slash",
			absPath: homeDir + string(filepath.Separator),
			want:    "~",
		},
	}

	if homeDir == "/" {
		tests = append(tests,
			struct {
				name    string
				absPath string
				want    string
			}{
				name:    "home is root, path is root",
				absPath: "/",
				want:    "~",
			},
			struct {
				name    string
				absPath string
				want    string
			}{
				name:    "home is root, path is /foo",
				absPath: "/foo",
				want:    filepath.Join("~", "foo"),
			},
		)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantNormalized := filepath.FromSlash(tt.want)
			got := toUserFriendlyPath(tt.absPath)
			gotNormalized := filepath.FromSlash(got)

			if gotNormalized != wantNormalized {
				t.Errorf("toUserFriendlyPath(%q) = %q, want %q", tt.absPath, gotNormalized, wantNormalized)
			}
		})
	}
}
