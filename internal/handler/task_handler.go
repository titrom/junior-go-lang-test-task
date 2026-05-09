package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"test_server/internal/model"
	"test_server/internal/storage"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func errorWrite(w http.ResponseWriter, statusCode int, err error) {
	writeJSON(w, statusCode, errorResponse{
		Err: err.Error(),
	})
}

func readBody(r *http.Request, v any) error {
	d, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if len(d) == 0 {
		return fmt.Errorf("body is empty")
	}

	err = json.Unmarshal(d, v)

	if err != nil {
		return err
	}

	return nil
}

type errorResponse struct {
	Err string `json:"error"`
}

func (er errorResponse) Error() string {
	return er.Err
}

type Handler struct {
	stor storage.Storage
}

func NewHandler(stor storage.Storage) *Handler {
	return &Handler{
		stor: stor,
	}
}

type task struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *Handler) TasksPostHandler(w http.ResponseWriter, r *http.Request) {
	var task = task{}
	err := readBody(r, &task)

	if err != nil {
		errorWrite(w, http.StatusBadRequest, err)
		return
	}

	log.Println(task)

	if strings.TrimSpace(task.Title) == "" {
		errorWrite(w, http.StatusBadRequest, fmt.Errorf("title is empty"))
		return
	}

	t, err := h.stor.Add(task.Title, task.Description)
	if err != nil {
		errorWrite(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) TasksAllGetHandler(w http.ResponseWriter, r *http.Request) {

	var tasks []model.Task

	status := model.TaskStatus(r.URL.Query().Get("status"))

	if status != "" {
		if !status.StatusCheck() {
			errorWrite(w, http.StatusBadRequest, fmt.Errorf("This status {%s} does not exist", status))
			return
		}
		tasks = h.stor.AllTasksByStatus(status)

	} else {
		tasks = h.stor.AllTasks()
	}

	log.Println(tasks)
	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) TasksByIdGetHandler(w http.ResponseWriter, r *http.Request) {
	s_id := r.PathValue("id")

	id, err := strconv.Atoi(s_id)

	if err != nil {
		errorWrite(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.stor.TaskById(uint(id))
	if err != nil {
		errorWrite(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) TasksUpdatePutHandler(w http.ResponseWriter, r *http.Request) {
	s_id := r.PathValue("id")

	id, err := strconv.Atoi(s_id)

	if err != nil {
		errorWrite(w, http.StatusBadRequest, err)
		return
	}

	var newTask model.UpdateTask
	if err = readBody(r, &newTask); err != nil {
		errorWrite(w, http.StatusBadRequest, err)
		return
	}

	if strings.TrimSpace(newTask.Title) == "" {
		errorWrite(w, http.StatusBadRequest, fmt.Errorf("title is empty"))
		return
	}

	if newTask.Status != nil {
		if !newTask.Status.StatusCheck() {
			errorWrite(w, http.StatusBadRequest, fmt.Errorf("This status {%s} does not exist", *newTask.Status))
			return
		}
	}

	task, err := h.stor.UpdateTask(uint(id), newTask)

	if err != nil {
		errorWrite(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) TaskDonePatchHandler(w http.ResponseWriter, r *http.Request) {
	s_id := r.PathValue("id")

	id, err := strconv.Atoi(s_id)

	if err != nil {
		errorWrite(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.stor.DoneTask(uint(id))

	if err != nil {
		errorWrite(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	s_id := r.PathValue("id")

	id, err := strconv.Atoi(s_id)

	if err != nil {
		errorWrite(w, http.StatusBadRequest, err)
		return
	}

	if err := h.stor.DeleteTask(uint(id)); err != nil {
		errorWrite(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
