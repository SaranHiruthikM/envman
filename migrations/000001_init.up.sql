-- Creates all envman tables in dependency order (parents before children)

-- 1. users (no dependencies)
CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        NOT NULL UNIQUE,
    name          TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMP   NOT NULL DEFAULT NOW()
);

-- 2. projects (depends on: users)
CREATE TABLE IF NOT EXISTS projects (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    created_by  UUID        NOT NULL REFERENCES users(id),
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW()
);

-- 3. user_projects (depends on: users, projects)
CREATE TABLE IF NOT EXISTS user_projects (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
project_id  UUID    NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    role        TEXT        NOT NULL
);

-- 4. secrets (depends on: projects, users)
CREATE TABLE IF NOT EXISTS secrets (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id     UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    pushed_by      UUID        NOT NULL REFERENCES users(id),
    encrypted_data TEXT        NOT NULL,
    pushed_at      TIMESTAMP   NOT NULL DEFAULT NOW()
);

-- 5. audit_logs (depends on: projects, users, secrets)
CREATE TABLE IF NOT EXISTS audit_logs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    pushed_by   UUID        NOT NULL REFERENCES users(id),
    secrets_id  UUID        NOT NULL REFERENCES secrets(id),
    pushed_at   TIMESTAMP   NOT NULL DEFAULT NOW()
);