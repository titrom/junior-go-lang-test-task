package model

type TaskStatus string

const (
	StatusNew        TaskStatus = "new"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

func (ts TaskStatus) StatusCheck() bool {
	return ts == StatusNew || ts == StatusInProgress || ts == StatusDone
}

type Task struct {
	Id          uint       `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

type UpdateTask struct {
	Title       string      `json:"title"`
	Description *string     `json:"description,omitempty"`
	Status      *TaskStatus `json:"status,omitempty"`
}
