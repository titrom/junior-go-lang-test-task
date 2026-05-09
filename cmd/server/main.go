package main

import (
	"fmt"
	"net/http"
	"test_server/internal/handler"
	"test_server/internal/storage"
	"time"
)

func main() {

	stor := storage.NewMemoryStorage()

	h := handler.NewHandler(stor)

	mux := http.NewServeMux()
	s := &http.Server{
		Addr:           ":6060",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	mux.HandleFunc("POST /tasks", h.TasksPostHandler)
	mux.HandleFunc("GET /tasks", h.TasksAllGetHandler)
	mux.HandleFunc("GET /tasks/{id}", h.TasksByIdGetHandler)
	mux.HandleFunc("PUT /tasks/{id}", h.TasksUpdatePutHandler)
	mux.HandleFunc("PATCH /tasks/{id}/done", h.TaskDonePatchHandler)
	mux.HandleFunc("DELETE /tasks/{id}", h.DeleteTask)

	if err := s.ListenAndServe(); err != nil {
		fmt.Println(err)
		return
	}
}
