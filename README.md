# Task Manager API

REST API для управления личными задачами, написан на Go с использованием только стандартной библиотеки.

## Запуск

```bash
go run ./cmd/server
```

Сервер запустится на порту `:6060`.

## Тесты

```bash
go test ./...
```

## API

### Создать задачу

```bash
POST /tasks
Content-Type: application/json

{"title": "Купить продукты", "description": "Молоко, хлеб, сыр"}
```

Ответ `201 Created`:

```json
{"id": 1, "title": "Купить продукты", "description": "Молоко, хлеб, сыр", "status": "new", "created_at": "2024-01-01T12:00:00Z", "updated_at": "2024-01-01T12:00:00Z"}
```

### Получить список задач

```bash
GET /tasks
GET /tasks?status=new       # фильтр по статусу
GET /tasks?status=in_progress
GET /tasks?status=done
```

### Получить задачу по ID

```bash
GET /tasks/{id}
```

### Обновить задачу

```bash
PUT /tasks/{id}
Content-Type: application/json

{"title": "Новое название", "description": "Новое описание", "status": "in_progress"}
```

### Завершить задачу

```bash
PATCH /tasks/{id}/done
```

### Удалить задачу

```bash
DELETE /tasks/{id}
```

Ответ `204 No Content`.

## Статусы задачи

| Статус | Описание |
|---|---|
| `new` | Новая задача |
| `in_progress` | В работе |
| `done` | Завершена |

## Формат ошибок

```json
{"error": "title is required"}
```

## Структура проекта

```
.
├── cmd/server/main.go          — точка входа
├── internal/
│   ├── handler/task_handler.go — HTTP-обработчики
│   ├── model/task.go           — модель задачи
│   └── storage/memory.go       — хранилище в памяти
└── go.mod
```
