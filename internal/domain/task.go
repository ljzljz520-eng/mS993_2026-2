package domain

type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskCompleted TaskStatus = "completed"
)

type Task struct {
	ID        string     `json:"id"`
	Kind      string     `json:"kind"`
	ProductID string     `json:"productId"`
	Delta     int        `json:"delta"`
	Status    TaskStatus `json:"status"`
}

type TaskRunStatus string

const (
	TaskRunExecuted TaskRunStatus = "executed"
	TaskRunAlready  TaskRunStatus = "already-processed"
	TaskRunFailed   TaskRunStatus = "failed"
)

type TaskRunResult struct {
	Status TaskRunStatus `json:"status"`
	TaskID string        `json:"taskId"`
	Worker string        `json:"worker"`
	Error  string        `json:"error,omitempty"`
}
