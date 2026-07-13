-- name: CreateTask :one
INSERT INTO tasks (title, description, status, project_id, assignee_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, title, description, status, project_id, assignee_id, created_at, updated_at;

-- name: GetTaskByID :one
SELECT id, title, description, status, project_id, assignee_id, created_at, updated_at
FROM tasks
WHERE id = $1;

-- name: GetTasksByProject :many
SELECT id, title, description, status, project_id, assignee_id, created_at, updated_at
FROM tasks
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: UpdateTask :one
UPDATE tasks
SET title = $2, description = $3, status = $4, assignee_id = $5, updated_at = NOW()
WHERE id = $1
RETURNING id, title, description, status, project_id, assignee_id, created_at, updated_at;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1;