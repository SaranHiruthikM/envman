-- Drops all envman tables in reverse dependency order (children before parents)

-- 5. audit_logs (depends on: projects, users, secrets)
DROP TABLE IF EXISTS audit_logs;

-- 4. secrets (depends on: projects, users)
DROP TABLE IF EXISTS secrets;

-- 3. user_projects (depends on: users, projects)
DROP TABLE IF EXISTS user_projects;

-- 2. projects (depends on: users)
DROP TABLE IF EXISTS projects;

-- 1. users (no dependencies)
DROP TABLE IF EXISTS users;