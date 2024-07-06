package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"schedulerapi"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

const (
	WebServerAddress = ":7540"
)

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

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web"))
	CheckDBFile()
	db, _ := sqlx.Connect("sqlite", "./scheduler.db")
	scheduler := schedulerapi.NewSchedulerStore(db)
	FileServer(r, "/", filesDir)
	r.Get("/api/nextdate", schedulerapi.NextDateHandler)
	r.Post("/api/task", scheduler.SetTaskHandler)
	r.Get("/api/task", scheduler.GetTaskHandler)
	r.Get("/api/tasks", scheduler.GetTasksHandler)
	r.Put("/api/task", scheduler.EditTaskHandler)
	r.Delete("/api/task", scheduler.DeleteTaskHandler)
	r.Post("/api/task/done", scheduler.DoneTaskHandler)
	log.Fatal(http.ListenAndServe(WebServerAddress, r))
}

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
