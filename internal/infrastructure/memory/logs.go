package memory

import (
	"sync"

	"aroma-maintenance/internal/domain"
)

type LogStore struct {
	mu   sync.RWMutex
	logs []domain.OperationLog
}

func NewLogStore() *LogStore {
	return &LogStore{logs: make([]domain.OperationLog, 0, 16)}
}

func (s *LogStore) Append(entry domain.OperationLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry.Sequence = len(s.logs) + 1
	s.logs = append(s.logs, entry)
}

func (s *LogStore) List() []domain.OperationLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]domain.OperationLog, len(s.logs))
	copy(result, s.logs)
	return result
}
