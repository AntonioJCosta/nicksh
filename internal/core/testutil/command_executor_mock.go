package testutil

import "errors"

// MockCommandExecutor is a mock implementation of ports.CommandExecutor.
type MockCommandExecutor struct {
	ExecuteFunc func(shellName, pipeline string) (stdout string, stderr string, err error)
}

// Execute calls the mock ExecuteFunc.
func (m *MockCommandExecutor) Execute(shellName, pipeline string) (string, string, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(shellName, pipeline)
	}
	return "", "", errors.New("MockCommandExecutor.ExecuteFunc not implemented")
}
