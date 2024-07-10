package main

import (
	controller "go_final_project_ver3/controller"
	service "go_final_project_ver3/service"
	storage "go_final_project_ver3/storage"
)

func main() {
	db, _ := storage.OpenSQLiteDB()
	defer db.Close()

	store := storage.NewSchedulerStore(db)
	service := service.NewTaskService(store)
	handler := controller.NewTaskHandler(service)
	controller.StartWebServer(handler)

}
