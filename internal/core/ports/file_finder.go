package ports

// HistoryFileFinder defines the contract for finding a history file.
type HistoryFileFinder interface {
	Find() (string, error)
}
