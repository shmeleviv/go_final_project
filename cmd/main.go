package main

import (
	"go_final_project_ver3/controller"
	"go_final_project_ver3/service"
	"go_final_project_ver3/storage"
)

func main() {
	db, _ := storage.OpenSQLiteDB()
	defer db.Close()

	store := storage.NewSchedulerStore(db)
	service := service.NewTaskService(store)
	handler := controller.NewTaskHandler(service)
	controller.StartWebServer(handler)

}
