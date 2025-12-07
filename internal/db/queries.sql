-- Deployments CRUD

-- name: CreateDeployment :one
INSERT INTO deployments (pkg_name, deployed_at)
VALUES (?, ?)
RETURNING id, pkg_name, deployed_at;

-- name: GetDeployment :one
SELECT id, pkg_name, deployed_at
FROM deployments
WHERE id = ?;

-- name: ListDeployments :many
SELECT id, pkg_name, deployed_at
FROM deployments
ORDER BY deployed_at DESC;

-- name: UpdateDeployment :one
UPDATE deployments
SET pkg_name = ?, deployed_at = ?
WHERE id = ?
RETURNING id, pkg_name, deployed_at;

-- name: DeleteDeployment :exec
DELETE FROM deployments
WHERE id = ?;

-- Executors CRUD

-- name: CreateExecutor :one
INSERT INTO executors (deployment_id, executor_name, container_id)
VALUES (?, ?, ?)
RETURNING id, deployment_id, executor_name, container_id;

-- name: GetExecutor :one
SELECT id, deployment_id, executor_name, container_id
FROM executors
WHERE id = ?;

-- name: ListExecutorsByDeployment :many
SELECT id, deployment_id, executor_name, container_id
FROM executors
WHERE deployment_id = ?
ORDER BY id;

-- name: UpdateExecutor :one
UPDATE executors
SET executor_name = ?, container_id = ?
WHERE id = ?
RETURNING id, deployment_id, executor_name, container_id;

-- name: DeleteExecutor :exec
DELETE FROM executors
WHERE id = ?;

-- Deployment with executors (join)

-- name: GetDeploymentWithExecutors :many
SELECT
    sqlc.embed(d) AS deployment,
    sqlc.embed(e) AS executor
FROM deployments d
LEFT JOIN executors e ON e.deployment_id = d.id
WHERE d.id = ?
ORDER BY e.id;
