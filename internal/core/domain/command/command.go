package command

// AnalyzedCommand holds the results of analyzing a command string.
type AnalyzedCommand struct {
	Original        string
	CommandName     string // The primary command/executable name
	IsComplex       bool
	PotentialArgs   []string // Arguments to the command, quotes stripped
	EffectiveLength int      // length of Original command without spaces

}
