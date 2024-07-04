package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type ResponseError struct {
	Data map[string]string // For extra error fields e.g. reason, details, etc.
}

type Task struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}
type TaskID struct {
	Id string `json:"id"`
}

type Tasks struct {
	//Id      int64  `json:"id"`
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func count(db *sqlx.DB) (int, error) {
	var count int
	return count, db.Get(&count, `SELECT count(id) FROM scheduler`)
}

func InstallDB() {
	file, err := os.Create("scheduler.db") // Create SQLite file
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	sqliteDatabase, err := sql.Open("sqlite", "./scheduler.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer sqliteDatabase.Close()
	createSchedulerTableSQL := `CREATE TABLE scheduler (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 date CHAR(8) NOT NULL DEFAULT "",
 title VARCHAR(64) NOT NULL DEFAULT "",
 comment VARCHAR(64) NOT NULL DEFAULT "",
 repeat VARCHAR(128) NOT NULL DEFAULT ""
);`
	createIndexScheduler := `CREATE INDEX scheduler_date ON scheduler (date);`

	statement, err := sqliteDatabase.Prepare(createSchedulerTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()                                                // Execute SQL Statements
	statement2, err := sqliteDatabase.Prepare(createIndexScheduler) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement2.Exec() // Execute SQL Statements

}

func CheckDBFile() {

	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true
	}
	// если install равен true, после открытия БД требуется выполнить
	// sql-запрос с CREATE TABLE и CREATE INDEX

	if install {
		InstallDB()

	}

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

func getNextDate(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "PANIC!!!", http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(nextDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//fmt.Println(fmt.Sprintf("%v", bytes.ReplaceAll(resp, []byte("\""), []byte(""))))
	w.Write(bytes.ReplaceAll(resp, []byte("\""), []byte("")))
}

func SetTask(w http.ResponseWriter, r *http.Request) {
	nowTime := time.Now()
	task := Task{}
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
	fmt.Printf("TESTTTTT date: %v, title: %v,comment: %v, repeat: %v ", task.Date, task.Title, task.Comment, task.Repeat)
	if task.Title == "" || len(task.Title) == 0 {
		respError := map[string]string{"error": "Не указан заголовок задачи"}
		resp, err := json.Marshal(respError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		//fmt.Println(fmt.Sprintf("%v", bytes.ReplaceAll(resp, []byte("\""), []byte(""))))
		w.Write(resp)
		return

	}
	dateTimeCheck, err := time.Parse("20060102", task.Date)
	fmt.Printf("checking dateTIMe: %v, taskDate:%v ", dateTimeCheck, task.Date)
	if err != nil && len(task.Date) != 0 && task.Date != "" {
		respError := map[string]string{"error": "Неправильный формат даты"}
		resp, err := json.Marshal(respError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	if len(task.Repeat) != 0 && task.Repeat != "" {
		fmt.Printf("date: %v, title: %v,comment: %v, repeat: %v ", task.Date, task.Title, task.Comment, task.Repeat)
		if len(task.Date) == 0 || task.Date == "" {
			task.Date = nowTime.Format("20060102")
		}
		if task.Date <= nowTime.Format("20060102") {
			task.Date, err = NextDate(nowTime, task.Date, task.Repeat)
			if err != nil {
				fmt.Printf("nextDate err: %v", err)
				respError := map[string]string{"error": "Ошибка функции nextDate"}
				resp, err := json.Marshal(respError)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusBadRequest)
				w.Write(resp)
				return
			}

		}
	}

	if len(task.Repeat) == 0 || task.Repeat == "" {
		if len(task.Date) == 0 || task.Date == "" || task.Date < nowTime.Format("20060102") {
			task.Date = nowTime.Format("20060102")
		}

	}

	sqliteDatabase, err := sql.Open("sqlite", "./scheduler.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer sqliteDatabase.Close()
	res, err := sqliteDatabase.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES ( :date, :title, :comment, :repeat)",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		return
	}
	taskid := TaskID{}
	id, err := res.LastInsertId()
	taskid.Id = strconv.Itoa(int(id))

	resp, err := json.Marshal(taskid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	//fmt.Println(fmt.Sprintf("%v", bytes.ReplaceAll(resp, []byte("\""), []byte(""))))
	w.Write(resp)
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks := []Tasks{}
	nowTime := time.Now()
	now := nowTime.Format("20060102")

	sqliteDatabase, err := sqlx.Connect("sqlite", "./scheduler.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer sqliteDatabase.Close()
	//err = sqliteDatabase.Get(&tasks, `SELECT * FROM scheduler WHERE date >= ? ORDER BY date LIMIT 50`, now)
	err = sqliteDatabase.Select(&tasks, `SELECT * FROM scheduler WHERE date >=? ORDER BY date LIMIT 50`, now)
	if err != nil {
		return
	}
	taskMap := map[string][]Tasks{"tasks": tasks}
	resp, err := json.Marshal(taskMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	// tasks := []Tasks{}
	getTask := Tasks{}
	id := r.FormValue("id")
	if len(id) == 0 {
		respError := map[string]string{"error": "Не указан идентификатор"}
		resp, err := json.Marshal(respError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}
	sqliteDatabase, err := sqlx.Connect("sqlite", "./scheduler.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer sqliteDatabase.Close()
	err = sqliteDatabase.Get(&getTask, "SELECT * FROM scheduler WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			respError := map[string]string{"error": "Задача не найдена"}
			resp, err := json.Marshal(respError)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(resp)
			return
		} else {
			fmt.Printf("db err: %v", err)
			fmt.Println("Get Task db aborted")
			return
		}
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

func EditTask(w http.ResponseWriter, r *http.Request) {
	task := Tasks{}
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

	if task.Title == "" || len(task.Title) == 0 {
		respError := map[string]string{"error": "Не указан заголовок задачи"}
		resp, err := json.Marshal(respError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return

	}

	if task.Id == "" || len(task.Id) == 0 {
		respError := map[string]string{"error": "Задача не найдена"}
		resp, err := json.Marshal(respError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return

	}

	if tid, err := strconv.Atoi(task.Id); err != nil || tid > 2147483647 {
		respError := map[string]string{"error": "Задача не найдена"}
		resp, err := json.Marshal(respError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return

	}
	nowTime := time.Now()
	nextDate, err := NextDate(nowTime, task.Date, task.Repeat)
	if err != nil {
		respError := map[string]string{"error": "Ошибка функции nextDate"}
		resp, err := json.Marshal(respError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}
	if len(task.Date) == 0 || task.Date == "" {
		nextDate, err = NextDate(nowTime, nowTime.Format("20060102"), task.Repeat)
	}
	dateTimeCheck, err := time.Parse("20060102", task.Date)
	fmt.Printf("checking dateTIMe: %v", dateTimeCheck)
	if err != nil && len(task.Date) != 0 {
		respError := map[string]string{"error": "Неправильный формат даты"}
		resp, err := json.Marshal(respError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	sqliteDatabase, err := sql.Open("sqlite", "./scheduler.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer sqliteDatabase.Close()
	_, err = sqliteDatabase.Exec("UPDATE scheduler SET date= :date, title= :title, comment= :comment, repeat= :repeat WHERE id = :id",
		sql.Named("date", nextDate),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
		sql.Named("id", task.Id))
	if err != nil {
		return
	}
	taskid := map[string]string{}
	resp, err := json.Marshal(taskid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	//fmt.Println(fmt.Sprintf("%v", bytes.ReplaceAll(resp, []byte("\""), []byte(""))))
	w.Write(resp)
}

func DoneTask(w http.ResponseWriter, r *http.Request) {
	getTask := Tasks{}
	id := r.FormValue("id")
	if len(id) == 0 {
		respError := map[string]string{"error": "Не указан идентификатор"}
		resp, err := json.Marshal(respError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	if tid, err := strconv.Atoi(id); err != nil || tid > 2147483647 {
		respError := map[string]string{"error": "Задача не найдена"}
		resp, err := json.Marshal(respError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return

	}
	sqliteDatabase, err := sqlx.Connect("sqlite", "./scheduler.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer sqliteDatabase.Close()
	err = sqliteDatabase.Get(&getTask, "SELECT * FROM scheduler WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			respError := map[string]string{"error": "Задача не найдена"}
			resp, err := json.Marshal(respError)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(resp)
			return
		} else {
			fmt.Printf("db err: %v", err)
			fmt.Println("Get Task db aborted")
			return
		}
	}
	if len(getTask.Repeat) == 0 || getTask.Repeat == "" {
		_, err = sqliteDatabase.Exec("DELETE FROM scheduler WHERE id = :id",
			sql.Named("id", getTask.Id))
		if err != nil {
			if err == sql.ErrNoRows {
				respError := map[string]string{"error": "Фэйл в удалении"}
				resp, err := json.Marshal(respError)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusBadRequest)
				w.Write(resp)
				return
			} else {
				fmt.Printf("db err: %v", err)
				fmt.Println("Get Task db aborted")
				return
			}
		}
		taskid := map[string]string{}
		resp, err := json.Marshal(taskid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		//fmt.Println(fmt.Sprintf("%v", bytes.ReplaceAll(resp, []byte("\""), []byte(""))))
		w.Write(resp)

	}

	nowTime := time.Now()
	nextDate, err := NextDate(nowTime, getTask.Date, getTask.Repeat)
	if err != nil {
		respError := map[string]string{"error": "Ошибка функции nextDate"}
		resp, err := json.Marshal(respError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}
	_, err = sqliteDatabase.Exec("UPDATE scheduler SET date= :date WHERE id = :id",
		sql.Named("date", nextDate),
		sql.Named("id", getTask.Id))
	if err != nil {
		if err == sql.ErrNoRows {
			respError := map[string]string{"error": "Задача не найдена"}
			resp, err := json.Marshal(respError)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(resp)
			return
		} else {
			fmt.Printf("db err: %v", err)
			fmt.Println("Get Task db aborted")
			return
		}

	}

	taskid := map[string]string{}
	resp, err := json.Marshal(taskid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	//fmt.Println(fmt.Sprintf("%v", bytes.ReplaceAll(resp, []byte("\""), []byte(""))))
	w.Write(resp)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Index handler

	// Create a route along /files that will serve contents from
	// the ./data/ folder.
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web"))
	CheckDBFile()
	FileServer(r, "/", filesDir)
	r.Get("/api/nextdate", getNextDate)
	r.Post("/api/task", SetTask)
	r.Get("/api/task", GetTask)
	r.Get("/api/tasks", GetTasks)
	r.Put("/api/task", EditTask)
	r.Post("/api/task/done", DoneTask)
	http.ListenAndServe(":7540", r)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
