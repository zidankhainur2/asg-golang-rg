package repository

import (
	"a21hc3NpZ25tZW50/db/filebased"
	"a21hc3NpZ25tZW50/model"
	"fmt"
)

type TaskRepository interface {
	Store(task *model.Task) error
	Update(task *model.Task) error
	Delete(id int) error
	GetByID(id int) (*model.Task, error)
	GetList() ([]model.Task, error)
	GetTaskCategory(id int) ([]model.TaskCategory, error)
}

type taskRepository struct {
	filebased *filebased.Data
}

func NewTaskRepo(filebasedDb *filebased.Data) *taskRepository {
	return &taskRepository{
		filebased: filebasedDb,
	}
}

func (t *taskRepository) Store(task *model.Task) error {
	t.filebased.StoreTask(*task)

	return nil
}

func (t *taskRepository) Update(task *model.Task) error {
	err := t.filebased.UpdateTask(task.ID, *task)
	if err != nil {
		return fmt.Errorf("failed to update task: %v", err)
	}
	return nil
}

func (t *taskRepository) Delete(id int) error {
	err := t.filebased.DeleteTask(id)
	if err != nil {
		return fmt.Errorf("record not found")
	}
	return nil
}

func (t *taskRepository) GetByID(id int) (*model.Task, error) {
	task, err := t.filebased.GetTaskByID(id)
	if err != nil {
		return nil, fmt.Errorf("record not found")
	}
	return task, nil
}

func (t *taskRepository) GetList() ([]model.Task, error) {
	tasks, err := t.filebased.GetTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks: %v", err)
	}
	return tasks, nil
}

func (t *taskRepository) GetTaskCategory(id int) ([]model.TaskCategory, error) {
	taskCategories, err := t.filebased.GetTaskListByCategory(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch task categories: %v", err)
	}
	return taskCategories, nil
}
