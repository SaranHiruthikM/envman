DROP INDEX IF EXISTS idx_user_projects_project_id;
DROP INDEX IF EXISTS idx_user_projects_user_id;
DROP INDEX IF EXISTS idx_audit_logs_project_pushed_at;
DROP INDEX IF EXISTS idx_secrets_project_pushed_at;

ALTER TABLE user_projects DROP CONSTRAINT IF EXISTS user_projects_role_check;
ALTER TABLE user_projects DROP CONSTRAINT IF EXISTS user_projects_user_project_unique;

ALTER TABLE audit_logs ALTER COLUMN pushed_at TYPE TIMESTAMP;
ALTER TABLE secrets    ALTER COLUMN pushed_at TYPE TIMESTAMP;
ALTER TABLE projects   ALTER COLUMN created_at TYPE TIMESTAMP;
ALTER TABLE users      ALTER COLUMN created_at TYPE TIMESTAMP;
