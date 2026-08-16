package memory

import (
	"sync"

	"aroma-maintenance/internal/domain"
)

type TaskStore struct {
	mu    sync.RWMutex
	tasks map[string]domain.Task
}

func NewTaskStore(fixture []domain.Task) *TaskStore {
	tasks := make(map[string]domain.Task, len(fixture))
	for _, task := range fixture {
		tasks[task.ID] = task
	}
	return &TaskStore{tasks: tasks}
}

func (s *TaskStore) Get(id string) (domain.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	return task, ok
}

func (s *TaskStore) Complete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok || task.Status == domain.TaskCompleted {
		return false
	}
	task.Status = domain.TaskCompleted
	s.tasks[id] = task
	return true
}

// Claim atomically moves a pending task into processing so only one worker can
// perform its side effects.
func (s *TaskStore) Claim(id string) (domain.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok || task.Status != domain.TaskPending {
		return domain.Task{}, false
	}
	task.Status = domain.TaskProcessing
	s.tasks[id] = task
	return task, true
}

// Release returns a claimed task to pending when its work fails, allowing a
// later worker to retry it.
func (s *TaskStore) Release(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok || task.Status != domain.TaskProcessing {
		return false
	}
	task.Status = domain.TaskPending
	s.tasks[id] = task
	return true
}
