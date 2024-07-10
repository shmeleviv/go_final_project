package service

import (
	entities "go_final_project_ver3/entities"
)

type TaskServiceINrerface interface {
	SetTask(task entities.SchedulerTask) error
	GetTask(id string) (entities.SchedulerTask, error)
	GetAllTasks() ([]entities.SchedulerTask, error)
	EditTask(task entities.SchedulerTask) error
	DoneTask(id string) error
	DeleteTask(id string) error
}
