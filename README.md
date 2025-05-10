# nicksh - Smart Shell Alias Manager

[![Go Report Card](https://goreportcard.com/badge/github.com/AntonioJCosta/nicksh)](https://goreportcard.com/report/github.com/AntonioJCosta/nicksh)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Latest Release](https://img.shields.io/github/v/release/AntonioJCosta/nicksh)](https://github.com/AntonioJCosta/nicksh/releases/latest)

Give your shell commands cool nicknames, and your fingers a break!

## Overview

`nicksh` is a command-line interface (CLI) tool built with Go that aims to streamline your shell experience by:

*   **Analyzing** your shell history to identify frequently used commands.
*   **Suggesting** concise and intuitive aliases for these commands.
*   **Interactively adding** suggested or predefined aliases to your shell configuration.
*   **Managing** aliases in a dedicated directory (`~/.nicksh/`) for easy sourcing.
*   **Leveraging `fzf`** (if available) for a powerful interactive selection experience, with fallback to numeric selection.

## Key Features

*   **Intelligent Alias Suggestions (`show`):** Scans your shell history (`.bash_history`, `.zsh_history`, etc.) to find commands you use often and suggests short, memorable aliases.
*   **Interactive Alias Addition (`add`):** After suggestions are shown, you can interactively select which aliases to add to your configuration using `fzf` or a numeric menu.
*   **Predefined Alias Management (`add-predefined`):** Add aliases from a curated `predefined_aliases.yaml` file. This is great for common commands or team-wide alias sets. You can interactively select which ones to add.
*   **List Managed Aliases (`list`):** View all aliases currently managed by `nicksh` in your `~/.nicksh/` directory.
*   **Safe Alias Generation:** Checks for conflicts with existing aliases and system commands before suggesting or adding new ones.
*   **Centralized Alias Files:** Stores generated aliases in `~/.nicksh/generated_aliases` (and potentially other files in `~/.nicksh/`), making it easy to source them into your shell.

## Installation

### 1. Using `go install` (Recommended)

If you have Go (1.24+) installed, you can install `nicksh` directly:

```bash
go install github.com/AntonioJCosta/nicksh/cmd/nicksh@latest
```
This will install the `nicksh` binary into your `$GOPATH/bin` or `$HOME/go/bin` directory. Ensure this directory is in your system's `PATH`.

### 2. From GitHub Releases 

You can download the latest pre-compiled binary for your operating system and architecture from the [GitHub Releases page](https://github.com/AntonioJCosta/nicksh/releases/latest).

1.  Go to the [Releases page](https://github.com/AntonioJCosta/nicksh/releases/latest).
2.  Download the appropriate archive for your system (e.g., `nicksh-linux-amd64.tar.gz`, `nicksh-windows-amd64.zip`).
3.  Extract the `nicksh` executable.
4.  Move the `nicksh` executable to a directory in your system's `PATH` (e.g., `/usr/local/bin` or `~/bin`).

   ```bash
   # Example for Linux:
   tar -xzf nicksh-linux-amd64.tar.gz
   sudo mv nicksh /usr/local/bin/
   ```

### 3. Manual Build from Source

1.  Clone the repository:
    ```bash
    git clone https://github.com/AntonioJCosta/nicksh.git
    cd nicksh
    ```
2.  Build the binary:
    ```bash
    go build -o nicksh ./cmd/nicksh/main.go
    ```
3.  Move the `nicksh` binary to a directory in your `PATH`:
    ```bash
    sudo mv nicksh /usr/local/bin/
    ```

## Setup: Sourcing `nicksh` Aliases

`nicksh` writes aliases to files within the `~/.nicksh/` directory (primarily `~/.nicksh/generated_aliases`). To make these aliases available in your shell, you need to source the files from this directory in your shell's configuration file (e.g., `~/.bashrc`, `~/.zshrc`, `~/.config/fish/config.fish`).

Add the following snippet to your shell configuration file:

```bash
# For bash/zsh
# Load all alias files from $HOME/.nicksh if the directory exists
if [ -d "$HOME/.nicksh" ]; then
  for file in "$HOME/.nicksh"/*; do
    [ -f "$file" ] && source "$file"
  done
fi
```

```fish
# For fish shell
# Load all alias files from $HOME/.nicksh if the directory exists
if test -d "$HOME/.nicksh"
    for file in "$HOME/.nicksh"/*
        if test -f "$file"
            source "$file"
        end
    end
end
```

After adding this, reload your shell configuration (e.g., `source ~/.bashrc`) or open a new terminal session.

## Usage Examples

### 1. Show Alias Suggestions: `nicksh show`

Analyzes your command history and suggests potential aliases.

```bash
nicksh show
```

**Flags:**

*   `--min-frequency, -f`: Minimum frequency for a command to be considered (default might be 3 or as defined in your code, e.g., `nicksh show -f 5`).
*   `--scan-limit, -s`: Number of recent history entries to scan (default might be 500, e.g., `nicksh show -s 1000`).
*   `--output-limit, -o`: Maximum number of suggestions to display.

```bash
# Show suggestions for commands used at least 5 times, scanning the last 1000 history entries
nicksh show -f 5 -s 1000
```
Output might look like:
```
Context: File: ~/.zsh_history
Suggested Aliases:
  alias gs='git status'
  alias gp='git push'
  alias ll='ls -alh'
(Source: Shell history analysis)
```

### 2. Interactively Add Suggested Aliases: `nicksh add`

This command typically follows `nicksh show` or can be run directly to process suggestions and add them. It will use `fzf` for selection if available, otherwise a numeric menu.

```bash
nicksh add
```
*(Assuming `add` command is implemented to take suggestions from `show` or re-trigger suggestion logic)*

If `fzf` is found:
```
Select aliases (TAB to multi-select, Enter to confirm) >
> alias gs='git status'
  alias gp='git push'
  alias ll='ls -alh'
```

If `fzf` is not found (numeric selection):
```
Select aliases to add (e.g., 1,3-5, or 'all', 'none'):
1. alias gs='git status'
2. alias gp='git push'
3. alias ll='ls -alh'
Enter selection: 1,3
```

### 3. Add Predefined Aliases: `nicksh add-predefined`

Interactively adds aliases from your `predefined_aliases.yaml` file.

```bash
nicksh add-predefined
```
This will present a list of valid aliases from your `predefined_aliases.yaml` file, allowing you to select which ones to add using `fzf` or numeric selection.

### 4. List Managed Aliases: `nicksh list`

Displays aliases currently managed by `nicksh` (found in `~/.nicksh/`).

```bash
nicksh list
```
Output:
```
Existing Aliases (managed by nicksh in $HOME/.nicksh/):
Note: These aliases are read from files in $HOME/.nicksh/.
      They reflect what nicksh manages, not necessarily your live shell's current alias state.
+------------+-----------------+
| ALIAS NAME | COMMAND         |
+------------+-----------------+
| gs         | git status      |
| gp         | git push        |
| gcm        | git commit -m   |
+------------+-----------------+
```

### Getting Help

For any command, you can use the `--help` flag to see available options:
```bash
nicksh --help
nicksh show --help
nicksh add-predefined --help
```

## Configuration

### Predefined Aliases (`predefined_aliases.yaml`)

`nicksh` came with already predefined aliases from a `predefined_aliases.yaml` file.

Example `predefined_aliases.yaml`:
```yaml
- name: gcm
  command: "git commit -m"
  description: "Git commit with message"
- name: kga
  command: "kubectl get all --all-namespaces"
  description: "Kubernetes get all resources in all namespaces"
```

## Contributing

Contributions are welcome! Whether it's reporting a bug, suggesting a feature, or submitting a pull request, your help is appreciated.

### Reporting Issues

Please use the [GitHub Issues](https://github.com/AntonioJCosta/nicksh/issues) tracker to report bugs or request features. Provide as much detail as possible, including:
*   Your operating system and shell.
*   The version of `nicksh` you are using (`nicksh --version`).
*   Steps to reproduce the bug.
*   Expected behavior and actual behavior.

### Pull Requests

1.  **Fork the repository** on GitHub.
2.  **Clone your fork** locally: `git clone https://github.com/YourUsername/nicksh.git`
3.  **Create a new branch** for your feature or bug fix: `git checkout -b my-feature-branch`
4.  **Make your changes.**
5.  **Add tests** for your changes if applicable.
6.  **Ensure tests pass:** `go test ./...`
7.  **Commit your changes:** `git commit -am "feat: Add some amazing feature"`
8.  **Push to your fork:** `git push origin my-feature-branch`
9.  **Open a Pull Request** on the `AntonioJCosta/nicksh` repository.

Please ensure your PR adheres to the existing code style and includes relevant tests. The `main` branch is protected, and PRs require review and passing checks (if configured).

### Development Setup

*   Go 1.24 or later.
*   To test interactive features locally, ensure `fzf` is installed and in your `PATH`.
*   Common Go development tools (linters, etc.) are recommended.

## License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.

## Acknowledgements

*   [Cobra](https://github.com/spf13/cobra) for CLI structure.
*   [fzf](https://github.com/junegunn/fzf) for the interactive fuzzy finder.
*   [color](https://github.com/fatih/color) for terminal colors.

---

Happy aliasing with `nicksh`!