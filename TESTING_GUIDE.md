# Как писать тесты для этого проекта

## Инструменты

Только стандартная библиотека — как и требует задание:

- `testing` — основной пакет для тестов
- `net/http/httptest` — поднимает фейковый HTTP-сервер в памяти, без реального порта
- `encoding/json` — разбирать ответы
- `bytes` / `strings` — формировать тело запроса

---

## Структура файлов

Тест кладётся рядом с тестируемым файлом, суффикс `_test.go`:

```
internal/
  handler/
    task_handler.go
    task_handler_test.go   ← тесты хендлеров
  storage/
    memory.go
    memory_test.go         ← тесты хранилища (опционально)
```

Пакет в тест-файле — тот же, что и у тестируемого кода:

```go
package handler
```

---

## Главный инструмент: httptest

`httptest.NewRecorder()` — заменяет `http.ResponseWriter`. После вызова хендлера можно читать код ответа и тело.

```go
req := httptest.NewRequest(http.MethodPost, "/tasks", body)
w   := httptest.NewRecorder()

handler.TasksPostHandler(w, req)

res := w.Result()
// res.StatusCode, res.Body — обычный http.Response
```

Никакого реального сервера не нужно. Тесты работают мгновенно.

---

## Шаблон одного теста

```go
func TestTasksPostHandler_Success(t *testing.T) {
    // 1. Подготовка
    stor := storage.NewMemoryStorage()
    h    := NewHandler(stor)

    body := strings.NewReader(`{"title":"Buy milk","description":"2 litres"}`)
    req  := httptest.NewRequest(http.MethodPost, "/tasks", body)
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    // 2. Вызов
    h.TasksPostHandler(w, req)

    // 3. Проверки
    res := w.Result()
    if res.StatusCode != http.StatusCreated {
        t.Errorf("want 201, got %d", res.StatusCode)
    }

    var task model.Task
    json.NewDecoder(res.Body).Decode(&task)

    if task.Title != "Buy milk" {
        t.Errorf("want title 'Buy milk', got '%s'", task.Title)
    }
    if task.Status != model.StatusNew {
        t.Errorf("want status 'new', got '%s'", task.Status)
    }
}
```

---

## Паттерн: таблица тестов (table-driven tests)

Стандартный Go-стиль — несколько сценариев в одной функции:

```go
func TestTasksPostHandler(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        wantStatus int
    }{
        {
            name:       "success",
            body:       `{"title":"Buy milk"}`,
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
            stor := storage.NewMemoryStorage()
            h    := NewHandler(stor)

            req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tc.body))
            w   := httptest.NewRecorder()

            h.TasksPostHandler(w, req)

            if w.Code != tc.wantStatus {
                t.Errorf("want %d, got %d", tc.wantStatus, w.Code)
            }
        })
    }
}
```

> `w.Code` — сокращение для `w.Result().StatusCode` у `httptest.ResponseRecorder`.

---

## Как тестировать эндпоинты с `{id}` в пути

`r.PathValue("id")` работает только если запрос прошёл через `ServeMux` с зарегистрированными маршрутами. При прямом вызове хендлера значение будет пустым.

**Решение** — создать мux в тесте и использовать `httptest.NewServer` или просто прогнать через `ServeHTTP`:

```go
func newTestMux() (*Handler, *http.ServeMux) {
    stor := storage.NewMemoryStorage()
    h    := NewHandler(stor)
    mux  := http.NewServeMux()

    mux.HandleFunc("POST /tasks",            h.TasksPostHandler)
    mux.HandleFunc("GET /tasks",             h.TasksAllGetHandler)
    mux.HandleFunc("GET /tasks/{id}",        h.TasksByIdGetHandler)
    mux.HandleFunc("PUT /tasks/{id}",        h.TashUpdatePutHandler)
    mux.HandleFunc("PATCH /tasks/{id}/done", h.TaskDonePatchHandler)
    mux.HandleFunc("DELETE /tasks/{id}",     h.DeleteTask)

    return h, mux
}

func TestGetTaskById_NotFound(t *testing.T) {
    _, mux := newTestMux()

    req := httptest.NewRequest(http.MethodGet, "/tasks/999", nil)
    w   := httptest.NewRecorder()

    mux.ServeHTTP(w, req)

    if w.Code != http.StatusNotFound {
        t.Errorf("want 404, got %d", w.Code)
    }
}
```

---

## Хелпер: создать задачу и получить её ID

Почти каждый тест на GET/PUT/DELETE требует сначала создать задачу. Вынеси в хелпер:

```go
func createTask(t *testing.T, mux http.Handler, title string) model.Task {
    t.Helper()

    body := fmt.Sprintf(`{"title":"%s"}`, title)
    req  := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
    w    := httptest.NewRecorder()

    mux.ServeHTTP(w, req)

    if w.Code != http.StatusCreated {
        t.Fatalf("createTask: want 201, got %d", w.Code)
    }

    var task model.Task
    json.NewDecoder(w.Body).Decode(&task)
    return task
}
```

Использование:

```go
func TestDeleteTask(t *testing.T) {
    _, mux := newTestMux()

    task := createTask(t, mux, "Test task")

    req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tasks/%d", task.Id), nil)
    w   := httptest.NewRecorder()
    mux.ServeHTTP(w, req)

    if w.Code != http.StatusNoContent {
        t.Errorf("want 204, got %d", w.Code)
    }
}
```

---

## Полный список тестов по заданию

| # | Тест | Эндпоинт | Что проверять |
|---|------|----------|---------------|
| 1 | Создание задачи | POST /tasks | статус 201, поля в ответе, status="new" |
| 2 | Ошибка без title | POST /tasks | статус 400, поле `error` в ответе |
| 3 | Получение по ID | GET /tasks/{id} | статус 200, совпадение полей |
| 4 | 404 для несуществующей | GET /tasks/{id} | статус 404 |
| 5 | Фильтр по status | GET /tasks?status=done | только задачи с нужным статусом |
| 6 | Обновление статуса | PUT /tasks/{id} | статус 200, обновлённые поля |
| 7 | Удаление задачи | DELETE /tasks/{id} | статус 204 |

---

## Запуск тестов

```bash
# все тесты
go test ./...

# с выводом каждого теста
go test ./... -v

# проверка гонок данных (обязательно для этого задания)
go test ./... -race

# конкретный тест
go test ./internal/handler/ -run TestTasksPostHandler -v
```

---

## Частые ошибки

**`r.PathValue` возвращает пустую строку** — хендлер вызван напрямую, не через mux. Используй `mux.ServeHTTP(w, req)`.

**Тест падает с `nil pointer`** — забыл инициализировать хранилище. Каждый тест должен создавать свой `storage.NewMemoryStorage()`.

**Тесты влияют друг на друга** — хранилище общее между тестами. Создавай новое хранилище в каждом тест-кейсе.

**`json.Decode` молча проглатывает ошибку** — всегда проверяй `err`:
```go
var task model.Task
if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
    t.Fatalf("decode response: %v", err)
}
```
