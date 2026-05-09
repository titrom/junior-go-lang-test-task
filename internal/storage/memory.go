package storage

import (
	"fmt"
	"sync"
	"test_server/internal/model"
	"time"
)

type Storage interface {
	Add(title string, description string) (model.Task, error)
	AllTasks() []model.Task
	AllTasksByStatus(model.TaskStatus) []model.Task
	TaskById(id uint) (model.Task, error)
	UpdateTask(id uint, newTask model.UpdateTask) (model.Task, error)
	DoneTask(id uint) (model.Task, error)
	DeleteTask(id uint) error
}

type memoryStorage struct {
	id    uint
	tasks map[uint]*model.Task
	mu    sync.Mutex
}

func NewMemoryStorage() *memoryStorage {
	return &memoryStorage{
		tasks: make(map[uint]*model.Task),
	}
}

func (ms *memoryStorage) Add(title string, description string) (model.Task, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.id += 1
	now := time.Now()
	task := &model.Task{
		Id:          ms.id,
		Title:       title,
		Description: description,
		Status:      model.StatusNew,
		CreatedAt:   now.Format(time.RFC3339),
		UpdatedAt:   now.Format(time.RFC3339),
	}
	ms.tasks[ms.id] = task
	return *task, nil
}

func (ms *memoryStorage) AllTasks() []model.Task {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	result := make([]model.Task, 0, len(ms.tasks))

	for _, v := range ms.tasks {
		result = append(result, *v)
	}

	return result
}

func (ms *memoryStorage) AllTasksByStatus(status model.TaskStatus) []model.Task {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	result := make([]model.Task, 0, len(ms.tasks))

	for _, v := range ms.tasks {
		if v.Status == status {
			result = append(result, *v)
		}
	}

	return result
}

func (ms *memoryStorage) checkId(id uint) bool {
	_, ok := ms.tasks[id]
	return ok
}

func (ms *memoryStorage) TaskById(id uint) (model.Task, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.checkId(id) {
		return model.Task{}, fmt.Errorf("task not found")
	}

	task := ms.tasks[id]

	return *task, nil
}

func (ms *memoryStorage) UpdateTask(id uint, newTask model.UpdateTask) (model.Task, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.checkId(id) {
		return model.Task{}, fmt.Errorf("task not found")
	}
	task := ms.tasks[id]

	task.Title = newTask.Title

	if newTask.Description != nil {
		task.Description = *newTask.Description
	}

	if newTask.Status != nil {
		task.Status = *newTask.Status
	}

	task.UpdatedAt = time.Now().Format(time.RFC3339)

	return *task, nil
}

func (ms *memoryStorage) DoneTask(id uint) (model.Task, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.checkId(id) {
		return model.Task{}, fmt.Errorf("task not found")
	}

	task := ms.tasks[id]

	task.Status = model.StatusDone

	task.UpdatedAt = time.Now().Format(time.RFC3339)

	return *task, nil

}

func (ms *memoryStorage) DeleteTask(id uint) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.checkId(id) {
		return fmt.Errorf("task not found")
	}

	delete(ms.tasks, id)
	return nil
}
