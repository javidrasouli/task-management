CREATE TABLE IF NOT EXISTS tasks (
    id          BIGSERIAL   PRIMARY KEY,
    title       TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    priority    TEXT        NOT NULL DEFAULT 'low',
    status      TEXT        NOT NULL DEFAULT 'todo',
    created_at  BIGINT      NOT NULL,
    updated_at  BIGINT      NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks (created_at DESC);
