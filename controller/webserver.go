package controller

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

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

func StartWebServer(handler TaskHandler) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web"))
	FileServer(r, "/", filesDir)
	r.Get("/api/nextdate", handler.NextDateHandler)
	r.Post("/api/task", handler.SetTaskHandler)
	r.Get("/api/task", handler.GetTaskHandler)
	r.Get("/api/tasks", handler.GetTasksHandler)
	r.Put("/api/task", handler.EditTaskHandler)
	r.Delete("/api/task", handler.DeleteTaskHandler)
	r.Post("/api/task/done", handler.DoneTaskHandler)
	log.Fatal(http.ListenAndServe(":7540", r))

}
