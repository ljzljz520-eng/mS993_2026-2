package application

import "aroma-maintenance/internal/domain"

type TaskRunner struct {
	repository     domain.TaskRepository
	perform        func(domain.Task, string) error
	BeforeComplete func(string)
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
	if err := r.perform(task, worker); err != nil {
		return domain.TaskRunResult{Status: domain.TaskRunFailed, TaskID: taskID, Worker: worker, Error: err.Error()}
	}
	if !r.repository.Complete(taskID) {
		return domain.TaskRunResult{Status: domain.TaskRunAlready, TaskID: taskID, Worker: worker}
	}
	return domain.TaskRunResult{Status: domain.TaskRunExecuted, TaskID: taskID, Worker: worker}
}
