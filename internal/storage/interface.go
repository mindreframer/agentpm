package storage

import "github.com/memomoo/agentpm/internal/epic"

type Storage interface {
	LoadEpic(filePath string) (*epic.Epic, error)
	SaveEpic(epic *epic.Epic, filePath string) error
	EpicExists(filePath string) bool
}

type Factory struct {
	useMemory bool
}

func NewFactory(useMemory bool) *Factory {
	return &Factory{useMemory: useMemory}
}

func (f *Factory) CreateStorage() Storage {
	if f.useMemory {
		return NewMemoryStorage()
	}
	return NewFileStorage()
}
