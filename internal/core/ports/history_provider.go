package ports

import "github.com/AntonioJCosta/nicksh/internal/core/domain/history"

type HistoryProvider interface {
	GetCommandFrequencies(scanLimit int, outputLimit int) ([]history.CommandFrequency, error)
	GetHistoryFilePath() string
	GetSourceIdentifier() string
}
