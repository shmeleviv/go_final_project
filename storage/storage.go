package storage

import (
	entities "go_final_project_ver3/entities"
)

type TaskRepository interface {
	GetAll() ([]entities.SchedulerTask, error)
	GetByID(id string) (entities.SchedulerTask, error)
	DeleteByID(id string) error
	UpdatebyID(task entities.SchedulerTask) error
}
