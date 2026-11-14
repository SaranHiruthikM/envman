-- Schema hardening: TZ-aware timestamps, missing constraints, hot-path indexes,
-- and ON DELETE policies so user deletion is actually possible.

ALTER TABLE users          ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';
ALTER TABLE projects       ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';
ALTER TABLE secrets        ALTER COLUMN pushed_at  TYPE TIMESTAMPTZ USING pushed_at  AT TIME ZONE 'UTC';
ALTER TABLE audit_logs     ALTER COLUMN pushed_at  TYPE TIMESTAMPTZ USING pushed_at  AT TIME ZONE 'UTC';

ALTER TABLE user_projects
    ADD CONSTRAINT user_projects_user_project_unique UNIQUE (user_id, project_id);

ALTER TABLE user_projects
    ADD CONSTRAINT user_projects_role_check CHECK (role IN ('owner', 'coordinator', 'member'));

-- Author FKs set NULL on user deletion so the record survives; container FKs cascade.
ALTER TABLE projects   DROP CONSTRAINT IF EXISTS projects_created_by_fkey;
ALTER TABLE projects
    ALTER COLUMN created_by DROP NOT NULL,
    ADD CONSTRAINT projects_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE secrets    DROP CONSTRAINT IF EXISTS secrets_pushed_by_fkey;
ALTER TABLE secrets
    ALTER COLUMN pushed_by DROP NOT NULL,
    ADD CONSTRAINT secrets_pushed_by_fkey
        FOREIGN KEY (pushed_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_pushed_by_fkey;
ALTER TABLE audit_logs
    ALTER COLUMN pushed_by DROP NOT NULL,
    ADD CONSTRAINT audit_logs_pushed_by_fkey
        FOREIGN KEY (pushed_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_secrets_id_fkey;
ALTER TABLE audit_logs
    ALTER COLUMN secrets_id DROP NOT NULL,
    ADD CONSTRAINT audit_logs_secrets_id_fkey
        FOREIGN KEY (secrets_id) REFERENCES secrets(id) ON DELETE SET NULL;

-- Hot-path indexes
CREATE INDEX IF NOT EXISTS idx_secrets_project_pushed_at
    ON secrets (project_id, pushed_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_project_pushed_at
    ON audit_logs (project_id, pushed_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_projects_user_id ON user_projects (user_id);
CREATE INDEX IF NOT EXISTS idx_user_projects_project_id ON user_projects (project_id);
