CREATE TABLE revisions (
    id           INTEGER PRIMARY KEY,
    package_name TEXT NOT NULL,
    version      TEXT NOT NULL,
    deploy_time  DATETIME NOT NULL
);

CREATE TABLE executors (
    id            INTEGER PRIMARY KEY,
    revision_id   INTEGER NOT NULL REFERENCES revisions(id) ON DELETE CASCADE,
    executor_name TEXT NOT NULL,
    anchor_name   TEXT NOT NULL,
    container_id  TEXT NOT NULL,
    img_name      TEXT NOT NULL,
    UNIQUE (revision_id, executor_name)
);
