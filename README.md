# envman

> Git for your `.env` files. Keep your team's secrets in sync — securely.

envman is a lightweight CLI tool that lets dev teams push and pull `.env` files without ever sharing secrets over Slack or WhatsApp. Secrets are encrypted on your machine before they ever leave it — your server never sees plaintext.

---

## How it works

```
Team lead runs:     envman push   →  encrypts .env locally → uploads to server
Teammate runs:      envman get    →  downloads → decrypts → shows diff → writes .env
```

One command to share. One command to sync. No dashboards, no per-seat pricing, no changing your workflow.

---

## Features

- **Client-side AES-256-GCM encryption** — secrets are encrypted before leaving your machine
- **Role-based access control** — owners manage the team, coordinators push secrets, members pull
- **Smart diff on get** — see exactly what changed before your `.env` is overwritten
- **Full audit trail** — every push is logged with who pushed and when
- **Zero workflow change** — your `.env` stays where it is, envman just keeps it in sync

---

## Requirements

- Go 1.21+
- Docker (for local Postgres)
- Git

---

## Installation

### From source

```bash
git clone https://github.com/SaranHiruthikM/envman
cd envman
go install ./cmd/envman/
```

Verify installation:

```bash
envman --help
```

---

## Server Setup (Self-hosted)

envman requires a backend server. You can run it locally or deploy it.

### 1. Clone the repo

```bash
git clone https://github.com/SaranHiruthikM/envman
cd envman
```

### 2. Start Postgres

```bash
docker run --name envman-db \
  -e POSTGRES_USER=envman \
  -e POSTGRES_PASSWORD=envman123 \
  -e POSTGRES_DB=envman \
  -p 5432:5432 \
  -d postgres:16
```

### 3. Configure environment

Create a `.env` file in the project root:

```env
DB_URL=postgres://envman:envman123@localhost:5432/envman
JWT_SECRET=your-super-secret-key-change-this
PORT=8080
```

### 4. Run the server

```bash
go run cmd/server/main.go
```

You should see:

```
Connected to database
Migrations applied
Server starting on port 8080
```

The server automatically runs database migrations on startup.

---

## CLI Usage

### Register & Login

```bash
# Create a new account
envman register

# Login to existing account
envman login

# Login against a custom server
envman login --server https://your-server.com
```

### Project Setup

Run this once inside your project folder to link it to envman:

```bash
cd your-project
envman init
```

This creates `.envman.json` in the current directory — **commit this file to git**. It contains your project ID and environment name (no secrets).

### Push & Get Secrets

```bash
# Push your local .env to the cloud (coordinators and owners only)
envman push

# Pull the latest .env from the cloud
envman get
```

`envman get` always shows a diff before overwriting:

```
Changes:
  + NEW_KEY=*** (new)
  ~ DATABASE_URL=*** (changed)
  - OLD_KEY (removed)

Apply changes? [y/n]:
```

### Team Management

```bash
# Invite a teammate (owners only)
envman invite

# Or with flags
envman invite --email teammate@example.com --role member

# Change a teammate's role (owners only)
envman role --user <user-id> --role coordinator
```

---

## Roles

| Role            | push | get | invite | change roles | delete project |
| --------------- | ---- | --- | ------ | ------------ | -------------- |
| **owner**       | ✅   | ✅  | ✅     | ✅           | ✅             |
| **coordinator** | ✅   | ✅  | ❌     | ❌           | ❌             |
| **member**      | ❌   | ✅  | ❌     | ❌           | ❌             |

The person who creates a project is automatically assigned the **owner** role.

---

## API Endpoints

| Method | Endpoint                        | Auth         | Description             |
| ------ | ------------------------------- | ------------ | ----------------------- |
| POST   | `/auth/register`                | Public       | Register a new user     |
| POST   | `/auth/login`                   | Public       | Login and get JWT token |
| POST   | `/projects`                     | Required     | Create a new project    |
| GET    | `/projects/:id`                 | Required     | Get project details     |
| POST   | `/projects/:id/members`         | Owner        | Invite a member         |
| PATCH  | `/projects/:id/members/:userId` | Owner        | Change member role      |
| DELETE | `/projects/:id/members/:userId` | Owner        | Remove a member         |
| POST   | `/projects/:id/secrets`         | Coordinator+ | Push encrypted secrets  |
| GET    | `/projects/:id/secrets`         | Member+      | Get latest secrets      |
| GET    | `/projects/:id/audit-logs`      | Member+      | View push history       |

---

## Security

- Passwords are hashed with **bcrypt** before storage
- Secrets are encrypted with **AES-256-GCM** on the client — the server only ever stores ciphertext
- JWT tokens expire after **30 days**
- Even if the database is compromised, secrets cannot be read without the user's token

---

## Project Structure

```
envman/
├── cmd/
│   ├── server/          # API server entry point
│   └── envman/          # CLI entry point
│       └── commands/    # cobra commands
├── internal/
│   ├── auth/            # register, login handlers
│   ├── projects/        # project handlers
│   ├── members/         # member management handlers
│   ├── secrets/         # secrets push/get handlers
│   ├── audit/           # audit log handlers
│   ├── middleware/       # JWT auth middleware
│   ├── permissions/      # role checking
│   ├── models/          # database structs
│   ├── db/              # database connection
│   ├── config/          # environment config
│   ├── token/           # JWT manager
│   ├── response/        # shared HTTP helpers
│   ├── common/          # shared DB helpers
│   └── cli/             # CLI config, HTTP client, crypto, diff
└── migrations/          # SQL migration files
```

---

## Configuration Files

**`.envman.json`** — lives in your project folder, committed to git:

```json
{
  "project_id": "abc-123-xyz",
  "env": "development"
}
```

**`~/.envman/config.json`** — lives on your machine, never committed:

```json
{
  "token": "your-jwt-token",
  "api_url": "http://localhost:8080"
}
```

---

## Roadmap

- [ ] Multiple environments (dev/staging/production)
- [ ] Auto-sync in background
- [ ] Email notifications on push
- [ ] Secret versioning and rollback
- [ ] `envman logs` command
- [ ] Homebrew / apt distribution

---

## Built With

- [Go](https://golang.org/) — CLI and API server
- [chi](https://github.com/go-chi/chi) — HTTP router
- [pgx](https://github.com/jackc/pgx) — PostgreSQL driver
- [cobra](https://github.com/spf13/cobra) — CLI framework
- [golang-migrate](https://github.com/golang-migrate/migrate) — database migrations
- [golang-jwt](https://github.com/golang-jwt/jwt) — JWT tokens
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) — password hashing

---

## License

MIT
