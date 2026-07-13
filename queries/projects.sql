-- name: CreateProject :one
INSERT INTO projects (name, description, owner_id)
VALUES ($1, $2, $3)
RETURNING id, name, description, owner_id, created_at, updated_at;

-- name: GetProjectByID :one
SELECT id, name, description, owner_id, created_at, updated_at
FROM projects
WHERE id = $1;

-- name: GetProjectsByUser :many
SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at, p.updated_at
FROM projects p
LEFT JOIN project_members pm ON pm.project_id = p.id
WHERE p.owner_id = $1 OR pm.user_id = $1;

-- name: UpdateProject :one
UPDATE projects
SET name = $2, description = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, name, description, owner_id, created_at, updated_at;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1;