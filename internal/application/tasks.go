package application

import (
	"sync"

	"aroma-maintenance/internal/domain"
)

type TaskRunner struct {
	repository     domain.TaskRepository
	perform        func(domain.Task, string) error
	BeforeComplete func(string)
	claimLocks     sync.Map
}

func NewTaskRunner(repository domain.TaskRepository, perform func(domain.Task, string) error) *TaskRunner {
	return &TaskRunner{repository: repository, perform: perform}
}

func (r *TaskRunner) Run(taskID, worker string) domain.TaskRunResult {
	task, ok := r.repository.Get(taskID)
	if !ok {
		return domain.TaskRunResult{Status: domain.TaskRunFailed, TaskID: taskID, Worker: worker, Error: "task not found"}
	}
	if task.Status == domain.TaskCompleted {
		return domain.TaskRunResult{Status: domain.TaskRunAlready, TaskID: taskID, Worker: worker}
	}
	if r.BeforeComplete != nil {
		r.BeforeComplete(worker)
	}
	unlock := func() {}
	if claimRepository, ok := r.repository.(domain.TaskClaimRepository); ok {
		claimedTask, claimed := claimRepository.Claim(taskID)
		if !claimed {
			return domain.TaskRunResult{Status: domain.TaskRunAlready, TaskID: taskID, Worker: worker}
		}
		task = claimedTask
		if err := r.perform(task, worker); err != nil {
			claimRepository.Release(taskID)
			return domain.TaskRunResult{Status: domain.TaskRunFailed, TaskID: taskID, Worker: worker, Error: err.Error()}
		}
	} else {
		// Preserve concurrency safety for repositories that predate the optional
		// atomic claim operation.
		lock := r.taskLock(taskID)
		lock.Lock()
		unlock = lock.Unlock
		task, ok = r.repository.Get(taskID)
		if !ok {
			unlock()
			return domain.TaskRunResult{Status: domain.TaskRunFailed, TaskID: taskID, Worker: worker, Error: "task not found"}
		}
		if task.Status == domain.TaskCompleted {
			unlock()
			return domain.TaskRunResult{Status: domain.TaskRunAlready, TaskID: taskID, Worker: worker}
		}
		if err := r.perform(task, worker); err != nil {
			unlock()
			return domain.TaskRunResult{Status: domain.TaskRunFailed, TaskID: taskID, Worker: worker, Error: err.Error()}
		}
	}
	if !r.repository.Complete(taskID) {
		unlock()
		return domain.TaskRunResult{Status: domain.TaskRunAlready, TaskID: taskID, Worker: worker}
	}
	unlock()
	return domain.TaskRunResult{Status: domain.TaskRunExecuted, TaskID: taskID, Worker: worker}
}

func (r *TaskRunner) taskLock(taskID string) *sync.Mutex {
	lock, _ := r.claimLocks.LoadOrStore(taskID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}
