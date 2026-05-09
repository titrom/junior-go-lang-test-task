package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"test_server/internal/model"
	"test_server/internal/storage"

	"testing"
)

func newTestMux() (*Handler, *http.ServeMux) {
	stor := storage.NewMemoryStorage()
	h := NewHandler(stor)
	mux := http.NewServeMux()

	mux.HandleFunc("POST /tasks", h.TasksPostHandler)
	mux.HandleFunc("GET /tasks", h.TasksAllGetHandler)
	mux.HandleFunc("GET /tasks/{id}", h.TasksByIdGetHandler)
	mux.HandleFunc("PUT /tasks/{id}", h.TasksUpdatePutHandler)
	mux.HandleFunc("PATCH /tasks/{id}/done", h.TaskDonePatchHandler)
	mux.HandleFunc("DELETE /tasks/{id}", h.DeleteTask)

	return h, mux
}

func createTask(t *testing.T, mux http.Handler, title string) model.Task {
	t.Helper()

	body := fmt.Sprintf(`{"title":"%s"}`, title)
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("createTask: want 201. got %d", w.Code)
	}

	var task model.Task
	json.NewDecoder(w.Body).Decode(&task)
	return task
}

type updateTask struct {
	id     uint
	status model.TaskStatus
}

func updateTaskForStatus(t *testing.T, mux http.Handler, update updateTask) model.Task {
	t.Helper()

	body, err := json.Marshal(model.UpdateTask{
		Title:  "Update status",
		Status: &update.status,
	})

	if err != nil {
		t.Errorf("%s %s", t.Name(), err.Error())
	}

	println(body)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/tasks/%d", update.id), bytes.NewReader(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update task: want 200. got %d", w.Code)
	}

	var task model.Task
	json.NewDecoder(w.Body).Decode(&task)
	return task
}

func TestTaskPostHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "success",
			body:       `{"title":"By milk"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "empty title",
			body:       `{"title":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, mux := newTestMux()

			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			if w.Code != tc.wantStatus {
				t.Errorf("want %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}

func TestTasksByIdGetHandler(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		wantStatus int
	}{
		{
			name:       "firs",
			id:         1,
			wantStatus: 200,
		}, {
			name:       "last",
			id:         1000,
			wantStatus: 200,
		}, {
			name:       "no found",
			id:         1001,
			wantStatus: 404,
		}, {
			name:       "not exist",
			id:         -1,
			wantStatus: 404,
		},
	}

	_, mux := newTestMux()

	for i := 0; i < 1000; i++ {
		createTask(t, mux, fmt.Sprintf("Test task %d", i))
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tasks/%d", tc.id), nil)

			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("want %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}

func TestTasksUpdatePutHandler(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		body       string
		wantStatus int
	}{
		{
			name:       "Test all",
			id:         1,
			body:       `{"title":"Test all", "descrition":"This is all test put", "status":"in_progress"}`,
			wantStatus: 200,
		}, {
			name:       "Test not status",
			id:         1,
			body:       `{"title":"est not status", "descrition":"This is not status test put"}`,
			wantStatus: 200,
		}, {
			name:       "Test not description",
			id:         1,
			body:       `{"title":"Test not description", "status":"done"}`,
			wantStatus: 200,
		}, {
			name:       "Test status not found",
			id:         1,
			body:       `{"title":"Test not description", "status":"test"}`,
			wantStatus: 400,
		},
		{
			name:       "Test not title",
			id:         1,
			body:       `{"descrition":"This is all test put", "status":"in_progress"}`,
			wantStatus: 400,
		},
		{
			name:       "Test no found",
			id:         2,
			body:       `{"title":"Test"}`,
			wantStatus: 404,
		},
		{
			name:       "Test empty body",
			id:         1,
			body:       `{}`,
			wantStatus: 400,
		},
	}

	_, mux := newTestMux()

	createTask(t, mux, "Test Put")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/tasks/%d", tc.id), strings.NewReader(tc.body))

			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("want %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}

func TestTaskDonePatchHandler(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		wantStatus int
	}{
		{
			name:       "success",
			id:         1,
			wantStatus: 200,
		}, {
			name:       "not found",
			id:         2,
			wantStatus: 404,
		},
	}

	_, mux := newTestMux()

	createTask(t, mux, "Test Patch")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/tasks/%d/done", tc.id), nil)

			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("want %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		wantStatus int
	}{
		{
			name:       "success",
			id:         1,
			wantStatus: 204,
		}, {
			name:       "not found",
			id:         2,
			wantStatus: 404,
		},
	}

	_, mux := newTestMux()

	createTask(t, mux, "Test Delete")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tasks/%d", tc.id), nil)

			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("want %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}

func TestTasksAllGetHandler(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		wantStatus int
	}{
		{
			name:       "Test get all",
			wantStatus: 200,
		},
		{
			name:       "Test get all for status new",
			status:     "new",
			wantStatus: 200,
		},
		{
			name:       "Test get all for status in_progress",
			status:     "in_progress",
			wantStatus: 200,
		}, {
			name:       "Test get all for status done",
			status:     "done",
			wantStatus: 200,
		},
		{
			name:       "Test get all for status not found",
			status:     "not_found",
			wantStatus: 400,
		},
	}

	_, mux := newTestMux()

	for i := 0; i < 1000; i++ {
		createTask(t, mux, fmt.Sprintf("Test task %d", i))
	}

	for i := 300; i < 401; i++ {
		updateTaskForStatus(t, mux, updateTask{
			id:     uint(i),
			status: model.StatusInProgress,
		})
	}
	for i := 500; i < 601; i++ {
		updateTaskForStatus(t, mux, updateTask{
			id:     uint(i),
			status: model.StatusDone,
		})
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			var req *http.Request

			if tc.status == "" {
				req = httptest.NewRequest(http.MethodGet, "/tasks", nil)

			} else {
				req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tasks?status=%s", tc.status), nil)
			}

			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("want %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}
