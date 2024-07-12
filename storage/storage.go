package storage

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"go_final_project_ver3/entities"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type SchedulerStore struct {
	DB *sqlx.DB
}

func NewSchedulerStore(db *sqlx.DB) SchedulerStore {
	database, _ := OpenSQLiteDB()
	return SchedulerStore{
		DB: database,
	}
}

func (r *SchedulerStore) GetAll(limit string) ([]entities.SchedulerTask, error) {
	tasks := []entities.SchedulerTask{}
	err := r.DB.Select(&tasks, `SELECT * FROM scheduler ORDER BY date LIMIT ?`, limit)
	if err != nil {
		return nil, errors.New("Ошибка получения задачи из БД")
	}
	return tasks, nil
}

func (r *SchedulerStore) GetByID(id string) (entities.SchedulerTask, error) {
	task := entities.SchedulerTask{}
	err := r.DB.Get(&task, "SELECT * FROM scheduler WHERE id = ?", id)
	if err != nil {
		return entities.SchedulerTask{}, errors.New("Ошибка получения задачи из БД")
	}
	return task, nil
}

func (r *SchedulerStore) DeleteByID(id string) error {
	getTask := entities.SchedulerTask{}
	err := r.DB.Get(&getTask, "SELECT * FROM scheduler WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Задача не найдена")
		}
		return errors.New("Ошибка получения задачи из БД")
	}
	_, err = r.DB.Exec("DELETE FROM scheduler WHERE id = :id",
		sql.Named("id", id))
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Задача не найдена")
		}
		return errors.New("Ошибка удаления задачи из БД")
	}
	return nil
}

func (r *SchedulerStore) Update(task entities.SchedulerTask) error {
	_, err := r.DB.Exec("UPDATE scheduler SET date= :date, title= :title, comment= :comment, repeat= :repeat WHERE id = :id",
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

func (r *SchedulerStore) Insert(task entities.SchedulerTask) (string, error) {

	res, err := r.DB.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES ( :date, :title, :comment, :repeat)",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		return "", errors.New("Ошибка добавления задачи в БД")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return "", err
	}
	// верните идентификатор последней добавленной записи
	return strconv.Itoa(int(id)), nil
}

func InstallDB() {
	file, err := os.Create("scheduler.db") // Create SQLite file
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	sqliteDatabase, err := sqlx.Connect("sqlite", "./scheduler.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer sqliteDatabase.Close()
	createSchedulerTableSQL := `CREATE TABLE scheduler (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 date CHAR(8) NOT NULL DEFAULT "",
 title VARCHAR(64) NOT NULL DEFAULT "",
 comment VARCHAR(512) NOT NULL DEFAULT "",
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

func OpenSQLiteDB() (*sqlx.DB, error) {

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
	db, err := sqlx.Connect("sqlite", "./scheduler.db")
	if err != nil {
		return nil, err
	}

	return db, nil
}
