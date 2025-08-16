package storage

import (
	"fmt"

	"github.com/mindreframer/agentpm/internal/epic"
)

type MemoryStorage struct {
	epics map[string]*epic.Epic
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		epics: make(map[string]*epic.Epic),
	}
}

func (ms *MemoryStorage) LoadEpic(filePath string) (*epic.Epic, error) {
	if e, exists := ms.epics[filePath]; exists {
		return e, nil
	}
	return nil, fmt.Errorf("epic file not found: %s", filePath)
}

func (ms *MemoryStorage) SaveEpic(e *epic.Epic, filePath string) error {
	if e == nil {
		return fmt.Errorf("epic cannot be nil")
	}
	ms.epics[filePath] = e
	return nil
}

func (ms *MemoryStorage) EpicExists(filePath string) bool {
	_, exists := ms.epics[filePath]
	return exists
}

func (ms *MemoryStorage) StoreEpic(filePath string, e *epic.Epic) {
	ms.epics[filePath] = e
}
