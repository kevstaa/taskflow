-- name: AddMember :one
INSERT INTO project_members (project_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING project_id, user_id, role, joined_at;

-- name: RemoveMember :exec
DELETE FROM project_members
WHERE project_id = $1 AND user_id = $2;

-- name: GetProjectMembers :many
SELECT u.id, u.username, u.email, pm.role, pm.joined_at
FROM project_members pm
JOIN users u ON u.id = pm.user_id
WHERE pm.project_id = $1;

-- name: IsMember :one
SELECT EXISTS (
    SELECT 1 FROM project_members
    WHERE project_id = $1 AND user_id = $2
) AS is_member;