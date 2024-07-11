package service

import (
	"errors"
	datetime "go_final_project_ver3/datetime"
	"go_final_project_ver3/entities"
	storage "go_final_project_ver3/storage"
	"strconv"
	"time"
)

const (
	GetAllTasksLimit = "50"
)

type TaskService struct {
	taskRepository storage.SchedulerStore
}

func NewTaskService(taskRepository storage.SchedulerStore) TaskService {
	return TaskService{
		taskRepository: taskRepository,
	}
}

func (s *TaskService) GetAllTasks() ([]entities.SchedulerTask, error) {
	nowTime := time.Now()
	now := nowTime.Format(datetime.DateFormat)
	tasks, err := s.taskRepository.GetAll(now, GetAllTasksLimit)
	if err != nil {
		return nil, errors.New("Ошибка получения задачи из БД")
	}
	return tasks, nil
}

func (s *TaskService) SetTask(task entities.SchedulerTask) (string, error) {
	nowTime := time.Now()
	if task.Title == "" || len(task.Title) == 0 {
		return "", errors.New("Не указан заголовок задачи")
	}

	_, err := time.Parse(datetime.DateFormat, task.Date)
	if err != nil && len(task.Date) != 0 && task.Date != "" {
		return "", errors.New("Неправильный формат даты")

	}

	if len(task.Repeat) != 0 && task.Repeat != "" {
		if len(task.Date) == 0 || task.Date == "" {
			task.Date = nowTime.Format(datetime.DateFormat)
		}
		if task.Date < nowTime.Format(datetime.DateFormat) {
			task.Date, err = datetime.NextDate(nowTime, task.Date, task.Repeat)
			if err != nil {
				return "", errors.New("Ошибка функции nextDate")
			}

		}
	}
	if len(task.Repeat) == 0 || task.Repeat == "" {
		if len(task.Date) == 0 || task.Date == "" || task.Date < nowTime.Format(datetime.DateFormat) {
			task.Date = nowTime.Format(datetime.DateFormat)
		}

	}

	id, err := s.taskRepository.Insert(task)
	if err != nil {
		return "", errors.New("Ошибка добавления задачи")
	}
	return id, nil

}

func (s *TaskService) GetTask(id string) (entities.SchedulerTask, error) {
	if len(id) == 0 {
		return entities.SchedulerTask{}, errors.New("Не указан идентификатор")
	}
	task, err := s.taskRepository.GetByID(id)
	if err != nil {
		return entities.SchedulerTask{}, errors.New("Ошибка получения задачи")
	}
	return task, nil

}

func (s *TaskService) EditTask(task entities.SchedulerTask) error {
	nowTime := time.Now()
	if task.Title == "" || len(task.Title) == 0 {
		return errors.New("Не указан заголовок задачи")
	}
	if task.Id == "" || len(task.Id) == 0 {
		return errors.New("Задача не найдена")
	}

	if tid, err := strconv.Atoi(task.Id); err != nil || tid > 2147483647 {
		return errors.New("Задача не найдена")
	}
	_, err := time.Parse(datetime.DateFormat, task.Date)
	if err != nil && len(task.Date) != 0 && task.Date != "" {
		return errors.New("Неправильный формат даты")

	}
	if len(task.Repeat) != 0 && task.Repeat != "" {
		if len(task.Date) == 0 || task.Date == "" {
			task.Date = nowTime.Format(datetime.DateFormat)
		}
		if task.Date <= nowTime.Format(datetime.DateFormat) {
			task.Date, err = datetime.NextDate(nowTime, task.Date, task.Repeat)
			if err != nil {
				return errors.New("Ошибка функции nextDate")
			}

		}
	}

	if len(task.Repeat) == 0 || task.Repeat == "" {
		if len(task.Date) == 0 || task.Date == "" || task.Date < nowTime.Format(datetime.DateFormat) {
			task.Date = nowTime.Format(datetime.DateFormat)
		}

	}

	err = s.taskRepository.Update(task)
	if err != nil {
		return errors.New("Ошибка обновления задачи")
	}

	return nil
}
func (s *TaskService) DoneTask(id string) error {
	getTask := entities.SchedulerTask{}
	if id == "" || len(id) == 0 {
		return errors.New("Не указан идентификатор")
	}

	if tid, err := strconv.Atoi(id); err != nil || tid > 2147483647 {
		return errors.New("Задача не найдена")
	}

	getTask, err := s.taskRepository.GetByID(id)
	if err != nil {
		return errors.New("Задача не найдена")
	}
	if len(getTask.Repeat) == 0 || getTask.Repeat == "" {

		//DELETE Task
		err = s.taskRepository.DeleteByID(getTask.Id)
		if err != nil {
			return errors.New("Ошибка завершения задачи")
		}
		return nil
	}
	nowTime := time.Now()
	nextDate, err := datetime.NextDate(nowTime, getTask.Date, getTask.Repeat)
	if err != nil {
		return errors.New("Ошибка функции NextDate")
	}
	getTask.Date = nextDate
	err = s.taskRepository.Update(getTask)
	if err != nil {
		return errors.New("Ошибка завершения задачи")
	}
	return nil
}

func (s *TaskService) DeleteTask(id string) error {
	getTask := entities.SchedulerTask{}
	if id == "" || len(id) == 0 {
		return errors.New("Не указан идентификатор")
	}

	if tid, err := strconv.Atoi(id); err != nil || tid > 2147483647 {
		return errors.New("Задача не найдена")
	}

	getTask, err := s.taskRepository.GetByID(id)
	if err != nil {
		return errors.New("Задача не найдена")
	}

	err = s.taskRepository.DeleteByID(getTask.Id)
	if err != nil {
		return errors.New("Ошибка удаления задачи")
	}
	return nil
}
