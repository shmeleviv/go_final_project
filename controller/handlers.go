package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go_final_project_ver3/datetime"
	"go_final_project_ver3/entities"
	"go_final_project_ver3/service"
)

type TaskHandler struct {
	taskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) TaskHandler {
	return TaskHandler{
		taskService: taskService,
	}
}

func (h *TaskHandler) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	getTask, err := h.taskService.GetTask(id)
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("%v", err)})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(resp)
		return
	}

	resp, err := json.Marshal(getTask)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

func (h *TaskHandler) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.taskService.GetAllTasks()
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("%v", err)})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(resp)
		return
	}

	resp, err := json.Marshal(map[string][]entities.SchedulerTask{"tasks": tasks})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)

}

func (h *TaskHandler) SetTaskHandler(w http.ResponseWriter, r *http.Request) {

	task := entities.SchedulerTask{}
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

	id, err := h.taskService.SetTask(task)
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("%v", err)})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(resp)
		return
	}

	resp, err := json.Marshal(map[string]string{"id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

func (h *TaskHandler) EditTaskHandler(w http.ResponseWriter, r *http.Request) {

	task := entities.SchedulerTask{}
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

	err = h.taskService.EditTask(task)
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("%v", err)})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(resp)
		return
	}

	resp, err := json.Marshal(map[string]string{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

func (h *TaskHandler) DoneTaskHandler(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")

	err := h.taskService.DoneTask(id)
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("%v", err)})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(resp)
		return
	}

	resp, err := json.Marshal(map[string]string{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

func (h *TaskHandler) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	err := h.taskService.DeleteTask(id)
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("%v", err)})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(resp)
		return
	}

	resp, err := json.Marshal(map[string]string{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

func (h *TaskHandler) NextDateHandler(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")
	nowTime, err := time.Parse(datetime.DateFormat, now)
	if err != nil {
		http.Error(w, "Incorrect now format", http.StatusBadRequest)
		return
	}
	nextDate, err := datetime.NextDate(nowTime, date, repeat)
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
	_, _ = w.Write(bytes.ReplaceAll(resp, []byte("\""), []byte("")))

}
