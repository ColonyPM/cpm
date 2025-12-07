CREATE TABLE deployments (
    id          INTEGER PRIMARY KEY,
    pkg_name    TEXT NOT NULL,
    deployed_at DATETIME NOT NULL
);

CREATE TABLE executors (
    id            INTEGER PRIMARY KEY,
    deployment_id INTEGER NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    executor_name TEXT NOT NULL,
    container_id  TEXT NOT NULL,
    -- Optional, but usually nice to avoid duplicates within a deployment
    UNIQUE (deployment_id, executor_name)
);
