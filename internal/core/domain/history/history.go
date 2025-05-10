/*
Package history defines core domain entities related to command history.
*/
package history

/*
CommandFrequency represents a command and its execution count.
This is a core domain entity.
*/
type CommandFrequency struct {
	Command string
	Count   int
}
