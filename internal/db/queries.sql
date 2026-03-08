-- Revisions CRUD

-- name: CreateRevision :one
INSERT INTO revisions (package_name, version, deploy_time)
VALUES (?, ?, ?)
RETURNING id, package_name, version, deploy_time;

-- name: GetRevision :one
SELECT id, package_name, version, deploy_time
FROM revisions
WHERE id = ?;

-- name: ListRevisions :many
SELECT id, package_name, version, deploy_time
FROM revisions
ORDER BY deploy_time DESC;

-- name: UpdateRevision :one
UPDATE revisions
SET package_name = ?, version = ?, deploy_time = ?
WHERE id = ?
RETURNING id, package_name, version, deploy_time;

-- name: DeleteRevision :exec
DELETE FROM revisions
WHERE id = ?;

-- name: GetRevisionWithExecutorsByPackageAndVersion :many
SELECT
    sqlc.embed(r),
    sqlc.embed(e)
FROM revisions r
LEFT JOIN executors e ON e.revision_id = r.id
WHERE r.package_name = ? AND r.version = ?
ORDER BY e.id;


-- Executors CRUD

-- name: CreateExecutor :one
INSERT INTO executors (revision_id, executor_name, anchor_name, container_id, img_name)
VALUES (?, ?, ?, ?, ?)
RETURNING id, revision_id, executor_name, anchor_name, container_id, img_name;

-- name: GetExecutor :one
SELECT id, revision_id, executor_name, anchor_name, container_id, img_name
FROM executors
WHERE id = ?;

-- name: ListExecutorsByRevision :many
SELECT id, revision_id, executor_name, anchor_name, container_id, img_name
FROM executors
WHERE revision_id = ?
ORDER BY id;

-- name: UpdateExecutor :one
UPDATE executors
SET executor_name = ?, anchor_name = ?, container_id = ?, img_name = ?
WHERE id = ?
RETURNING id, revision_id, executor_name, anchor_name, container_id, img_name;

-- name: DeleteExecutor :exec
DELETE FROM executors
WHERE id = ?;


-- Revision with executors

-- name: GetRevisionWithExecutors :many
SELECT
    sqlc.embed(r),
    sqlc.embed(e)
FROM revisions r
LEFT JOIN executors e ON e.revision_id = r.id
WHERE r.id = ?
ORDER BY e.id;


-- name: GetRevisionByPackageAndVersion :one
SELECT id, package_name, version, deploy_time
FROM revisions
WHERE package_name = ? AND version = ?;
