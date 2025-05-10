package ports

// CommandExecutor defines an interface for executing shell commands.
type CommandExecutor interface {
	Execute(shellName, pipeline string) (stdout string, stderr string, err error)
}
