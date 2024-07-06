package schedulerapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

const (
	GetTasksLimit = 50
)

type SchedulerTask struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type SchedulerStore struct {
	DB *sqlx.DB
}

func NewSchedulerStore(db *sqlx.DB) SchedulerStore {
	return SchedulerStore{DB: db}
}

func (s SchedulerStore) SetTaskHandler(w http.ResponseWriter, r *http.Request) {
	task := SchedulerTask{}
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(buf.Bytes(), &task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.SetTask(task)
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("%v", err)})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	resp, err := json.Marshal(map[string]string{"id": strconv.Itoa(id)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// Добавляем задачу,  в случае успеха возвращается id,  в случае неуспеха - ошибка
func (s SchedulerStore) SetTask(task SchedulerTask) (int, error) {
	nowTime := time.Now()
	if task.Title == "" || len(task.Title) == 0 {
		return 0, errors.New("Не указан заголовок задачи")
	}

	_, err := time.Parse("20060102", task.Date)
	if err != nil && len(task.Date) != 0 && task.Date != "" {
		return 0, errors.New("Неправильный формат даты")

	}

	if len(task.Repeat) != 0 && task.Repeat != "" {
		if len(task.Date) == 0 || task.Date == "" {
			task.Date = nowTime.Format("20060102")
		}
		if task.Date < nowTime.Format("20060102") {
			task.Date, err = NextDate(nowTime, task.Date, task.Repeat)
			if err != nil {
				return 0, errors.New("Ошибка функции nextDate")
			}

		}
	}
	if len(task.Repeat) == 0 || task.Repeat == "" {
		if len(task.Date) == 0 || task.Date == "" || task.Date < nowTime.Format("20060102") {
			task.Date = nowTime.Format("20060102")
		}

	}

	res, err := s.DB.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES ( :date, :title, :comment, :repeat)",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		return 0, errors.New("Ошибка добавления задачи в БД")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	// верните идентификатор последней добавленной записи
	return int(id), nil
}

func NextDate(now time.Time, date string, repeat string) (string, error) {

	//if len(repeat) == 0 || repeat == "" {
	//	return "repeat is empty format", errors.New("repeat is empty format")
	//}
	repeatSlice := strings.Split(repeat, " ")
	if repeatSlice[0] == "d" && len(repeatSlice) < 2 {
		return "incorrect d repeat format", errors.New("incorrect d repeat format")
	}
	if repeatSlice[0] == "y" && len(repeatSlice) > 1 {
		return "incorrect y repeat format", errors.New("incorrect y repeat format")
	}
	switch repeatSlice[0] {
	case "d":
		switch repeatSlice[1] {
		default:
			dayCount, err := strconv.Atoi(repeatSlice[1])
			if err != nil {
				return "Can't pasre interger", err
			}
			if dayCount >= 1 && dayCount <= 400 {

				if date == now.Format("20060102") {
					dateTime, err := time.Parse("20060102", date)
					if err != nil {
						return "Incorrect date format", err
					}
					if dayCount == 1 {
						//dateTime = dateTime.AddDate(0, 0, dayCount)
						return fmt.Sprintf("%v", dateTime.Format("20060102")), nil
					}
					dateTime = dateTime.AddDate(0, 0, dayCount)
					return fmt.Sprintf("%v", dateTime.Format("20060102")), nil
				}

				if date > now.Format("20060102") {
					fmt.Println("Case date is bigger than now date")
					dateTime, err := time.Parse("20060102", date)
					if err != nil {
						return "Incorrect date format", err
					}
					dateTime = dateTime.AddDate(0, 0, dayCount)
					return fmt.Sprintf("%v", dateTime.Format("20060102")), nil
				}

				if date < now.Format("20060102") {
					dateTime, err := time.Parse("20060102", date)
					if err != nil {
						return "Incorrect date format", err
					}
					fmt.Println("Case date is lower than now date")
					for {
						dateTime = dateTime.AddDate(0, 0, dayCount)
						if dateTime.Format("20060102") < now.Format("20060102") {
							continue
						} else {
							return fmt.Sprintf("%v", dateTime.Format("20060102")), nil
						}

					}
				}

			}
			return "Incorrect d format. d must be in rage 1 to 400", errors.New("Incorrect d format. d must be in rage 1 to 400")
		}

	case "y":
		fmt.Println("Case y works")
		//newTime := now.AddDate(1, 0, 0)
		dateTime, err := time.Parse("20060102", date)
		if err != nil {
			return "Incorrect date format", err
		}

		if date < now.Format("20060102") {
			dateTime, err := time.Parse("20060102", date)
			if err != nil {
				return "Incorrect date format", err
			}
			fmt.Println("Case date is lower than now date")
			for {
				dateTime = dateTime.AddDate(1, 0, 0)
				if dateTime.Format("20060102") < now.Format("20060102") {
					continue
				} else {
					return fmt.Sprintf("%v", dateTime.Format("20060102")), nil
				}

			}
		}
		dateTime = dateTime.AddDate(1, 0, 0)
		return fmt.Sprintf("%v", dateTime.Format("20060102")), nil
	default:

		return "Incorrect Format. Only d or y supporter", errors.New("Incorrect Format. Only d or y supporter")

	}
}

func (s SchedulerStore) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	getTask, err := s.GetTask(id)
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("%v", err)})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	resp, err := json.Marshal(getTask)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (s SchedulerStore) GetTask(id string) (SchedulerTask, error) {
	if len(id) == 0 {
		return SchedulerTask{}, errors.New("Не указан идентификатор")
	}
	task := SchedulerTask{}
	err := s.DB.Get(&task, "SELECT * FROM scheduler WHERE id = ?", id)
	if err != nil {
		return SchedulerTask{}, errors.New("Ошибка получения задачи из БД")
	}
	return task, nil
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")
	nowTime, err := time.Parse("20060102", now)
	if err != nil {
		http.Error(w, "Incorrect now format", http.StatusBadRequest)
		return
	}
	nextDate, err := NextDate(nowTime, date, repeat)
	if err != nil {
		http.Error(w, "NextDate failure", http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(nextDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes.ReplaceAll(resp, []byte("\""), []byte("")))
}

func (s SchedulerStore) GetTasks() ([]SchedulerTask, error) {
	nowTime := time.Now()
	now := nowTime.Format("20060102")
	tasks := []SchedulerTask{}
	err := s.DB.Select(&tasks, `SELECT * FROM scheduler WHERE date >=? ORDER BY date LIMIT ?`, now, GetTasksLimit)
	if err != nil {
		return nil, errors.New("Ошибка получения задачи из БД")
	}
	return tasks, nil
}

func (s SchedulerStore) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.GetTasks()
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("%v", err)})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	resp, err := json.Marshal(map[string][]SchedulerTask{"tasks": tasks})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (s SchedulerStore) EditTask(task SchedulerTask) error {
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

	_, err := time.Parse("20060102", task.Date)
	if err != nil && len(task.Date) != 0 && task.Date != "" {
		return errors.New("Неправильный формат даты")

	}

	if len(task.Repeat) != 0 && task.Repeat != "" {
		if len(task.Date) == 0 || task.Date == "" {
			task.Date = nowTime.Format("20060102")
		}
		if task.Date <= nowTime.Format("20060102") {
			task.Date, err = NextDate(nowTime, task.Date, task.Repeat)
			if err != nil {
				return errors.New("Ошибка функции nextDate")
			}

		}
	}
	if len(task.Repeat) == 0 || task.Repeat == "" {
		if len(task.Date) == 0 || task.Date == "" || task.Date < nowTime.Format("20060102") {
			task.Date = nowTime.Format("20060102")
		}

	}

	_, err = s.DB.Exec("UPDATE scheduler SET date= :date, title= :title, comment= :comment, repeat= :repeat WHERE id = :id",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
		sql.Named("id", task.Id))
	if err != nil {
		return errors.New("Ошибка обновления задачи в БД")
	}

	return nil
}

func (s SchedulerStore) EditTaskHandler(w http.ResponseWriter, r *http.Request) {
	task := SchedulerTask{}
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(buf.Bytes(), &task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.EditTask(task)
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("%v", err)})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	resp, err := json.Marshal(map[string]string{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (s SchedulerStore) DoneTask(id string) error {
	getTask := SchedulerTask{}
	if id == "" || len(id) == 0 {
		return errors.New("Не указан идентификатор")
	}

	if tid, err := strconv.Atoi(id); err != nil || tid > 2147483647 {
		return errors.New("Задача не найдена")
	}

	err := s.DB.Get(&getTask, "SELECT * FROM scheduler WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Задача не найдена")
		}
		return errors.New("Ошибка получения задачи из БД")
	}

	if len(getTask.Repeat) == 0 || getTask.Repeat == "" {

		//DELETE Task
		_, err = s.DB.Exec("DELETE FROM scheduler WHERE id = :id",
			sql.Named("id", getTask.Id))
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.New("Задача не найдена")
			}
			return errors.New("Ошибка удаления задачи из БД")
		}
		return nil
	}

	nowTime := time.Now()
	nextDate, err := NextDate(nowTime, getTask.Date, getTask.Repeat)
	if err != nil {
		return errors.New("Ошибка функции NextDate")
	}
	_, err = s.DB.Exec("UPDATE scheduler SET date= :date WHERE id = :id",
		sql.Named("date", nextDate),
		sql.Named("id", getTask.Id))
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Задача не найдена")
		}
		return errors.New("Ошибка обновления задачи в БД")
	}

	return nil
}

func (s SchedulerStore) DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	err := s.DoneTask(id)
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("%v", err)})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	resp, err := json.Marshal(map[string]string{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (s SchedulerStore) DeleteTask(id string) error {
	getTask := SchedulerTask{}
	if id == "" || len(id) == 0 {
		return errors.New("Не указан идентификатор")
	}

	if tid, err := strconv.Atoi(id); err != nil || tid > 2147483647 {
		return errors.New("Задача не найдена")
	}

	err := s.DB.Get(&getTask, "SELECT * FROM scheduler WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Задача не найдена")
		}
		return errors.New("Ошибка получения задачи из БД")
	}
	//DELETE Task
	_, err = s.DB.Exec("DELETE FROM scheduler WHERE id = :id",
		sql.Named("id", id))
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Задача не найдена")
		}
		return errors.New("Ошибка удаления задачи из БД")
	}
	return nil
}

func (s SchedulerStore) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	err := s.DeleteTask(id)
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("%v", err)})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	resp, err := json.Marshal(map[string]string{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
