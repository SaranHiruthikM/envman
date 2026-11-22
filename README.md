# envman (OFFCAMPUS CODE: V8V8W5)

> Git for your `.env` files. End-to-end encrypted. Your server never sees plaintext.

envman is a CLI tool that lets dev teams push and pull `.env` files without sharing secrets over Slack or WhatsApp. Secrets are encrypted on your machine with a per-project key that only you and your teammates hold — the server literally cannot decrypt them.

---

## How the crypto actually works

```
Register:   you generate an X25519 keypair locally.
            Your private key is encrypted with argon2id(password) and uploaded sealed.
            The server stores only the public key and the sealed blob.

Init:       the CLI generates a random 32-byte project key, wraps it for you
            (ECDH + HKDF + XChaCha20-Poly1305), and uploads the wrapped copy.

Invite:     owner fetches the invitee's public key, wraps the project key for them,
            and uploads the new wrapped copy. The plaintext project key never leaves
            the owner's machine.

Push:       CLI unwraps the project key, AES-256-GCM encrypts the .env body, uploads.
Get:        CLI unwraps the project key, downloads the ciphertext, decrypts locally.
```

A database dump reveals: your public key, a sealed private-key blob (needs your password to open), ciphertext bytes, and wrapped project keys (each needs that user's private key to open). It does **not** reveal anything you've pushed.

---

## Features

- **Client-side E2E encryption** — X25519 + HKDF + XChaCha20-Poly1305 for key wrapping, AES-256-GCM for the body, argon2id for the password KDF
- **Role-based access** — owner, coordinator, member
- **Smart diff on `get`** — see what's changing before your `.env` is overwritten; previous content saved to `.env.bak`
- **Audit trail** — every push logged
- **Non-interactive mode** — `envman get --yes` + `ENVMAN_TOKEN` env for CI
- **Single static Go binary** — Linux, macOS, Windows, amd64 + arm64

---

## Quick start (docker-compose)

```bash
# 1. Clone & configure
git clone https://github.com/SaranHiruthikM/envman
cd envman
cp .env.example .env
# Set a real JWT_SECRET in .env (needs 32+ chars):
#   openssl rand -base64 48 >> .env

# 2. Start Postgres + migrator + server
docker compose up -d

# 3. Install the CLI
go install ./cmd/envman

# 4. Use it
envman register
envman init
envman push
```

---

## Requirements

- Go 1.25+
- Docker (for local Postgres) or a managed Postgres 14+
- Git

---

## CLI usage

### Account

```bash
envman register               # create an account
envman login                  # login on a fresh machine (re-derives your private key)
envman logout                 # clear the stored session
envman version
```

### Project setup

```bash
cd your-project
envman init                   # generates project key, wraps for you, writes .envman.json
```

Commit `.envman.json` to git — it contains only the project UUID.

### Sync

```bash
envman push                   # encrypt and upload (coordinators + owners only)
envman get                    # download, decrypt, show diff, prompt before overwrite
envman get --yes              # non-interactive (CI); writes .env.bak before overwriting
```

### Team management

```bash
envman invite --email teammate@example.com --role member
envman role --user <user-id> --role coordinator
```

An invite requires the invitee to have already run `envman register` so the owner can fetch their public key.

---

## Roles

| Role            | push | get | invite | change roles | delete project |
| --------------- | ---- | --- | ------ | ------------ | -------------- |
| **owner**       | yes  | yes | yes    | yes          | yes            |
| **coordinator** | yes  | yes | no     | no           | no             |
| **member**      | no   | yes | no     | no           | no             |

Project creator is automatically owner. The last-owner invariant is enforced: you can't demote or remove the only owner.

---

## Deploy (self-host)

### Local via docker-compose

See **Quick start** above. The compose stack runs Postgres, a one-shot migrator, and the server behind a healthcheck.

### Production

The server is a single static binary (`cmd/server`) plus a one-shot migrator (`cmd/migrate`). Any platform that runs containers works.

**Recommended: Fly.io.** Deploy roughly:

```bash
fly launch --no-deploy
fly secrets set JWT_SECRET="$(openssl rand -base64 48)"
fly postgres attach <your-pg-cluster>
# Set release_command = "/usr/local/bin/envman-migrate up" in fly.toml
fly deploy
```

Expected cost at v1: ~$5–15/mo for the app + $5–10/mo for Fly Postgres.

**Environment variables:**

| Name           | Required | Notes                                                 |
| -------------- | -------- | ----------------------------------------------------- |
| `DB_URL`       | yes      | `postgres://user:pass@host:5432/dbname?sslmode=require` |
| `JWT_SECRET`   | yes      | 32+ characters; server refuses to boot otherwise      |
| `PORT`         | no       | default `8080`                                        |
| `RUN_MIGRATIONS` | no     | `true` to auto-migrate on boot (dev only — use `envman-migrate` in prod) |

---

## Config file locations

| File                       | Purpose                       | Committed? | Permissions |
| -------------------------- | ----------------------------- | ---------- | ----------- |
| `~/.envman/config.json`    | token + user's decrypted priv | no         | 0600        |
| `./.envman.json`           | project UUID                  | **yes**    | 0644        |
| `./.env`                   | your secrets                  | no         | 0600        |
| `./.env.bak`               | previous `.env` after `get`   | no         | 0600        |

---

## API endpoints

| Method | Endpoint                                    | Auth         | Description                       |
| ------ | ------------------------------------------- | ------------ | --------------------------------- |
| POST   | `/auth/register`                            | public       | Register (uploads sealed privkey) |
| POST   | `/auth/login`                               | public       | Login (returns sealed privkey)    |
| GET    | `/users/by-email/{email}/public-key`        | required     | Fetch pubkey for invite           |
| POST   | `/projects`                                 | required     | Create project                    |
| GET    | `/projects/{projectID}`                     | required     | Project details                   |
| POST   | `/projects/{projectID}/keys`                | required     | Store caller's wrapped project key |
| GET    | `/projects/{projectID}/keys/me`             | required     | Fetch caller's wrapped project key |
| POST   | `/projects/{projectID}/members`             | owner        | Invite (carries wrapped key)      |
| PATCH  | `/projects/{projectID}/members/{memberUserID}` | owner      | Change role                       |
| DELETE | `/projects/{projectID}/members/{memberUserID}` | owner      | Remove member                     |
| POST   | `/projects/{projectID}/secrets`             | coordinator+ | Push ciphertext                   |
| GET    | `/projects/{projectID}/secrets`             | member+      | Fetch latest ciphertext           |
| GET    | `/projects/{projectID}/audit-logs`          | member+      | Push history                      |
| GET    | `/healthz`                                  | public       | Liveness                          |
| GET    | `/readyz`                                   | public       | Readiness (pings DB)              |

All responses use a `{"data": ...}` envelope or `{"error": {"code", "message"}}` on failure.

---

## Security

See [`SECURITY.md`](./SECURITY.md) for the vulnerability reporting process and threat model.

- Passwords: bcrypt (cost 10), with a constant-time dummy compare on unknown emails to kill the enumeration oracle
- JWTs: HS256 pinned (no algorithm confusion), 30-day expiry, `JWT_SECRET` must be ≥32 chars
- Transport: TLS at the edge
- Database dump: ciphertext only — no key material usable without a client machine
- Audit: every push logged with user, time, revision

---

## Project structure

```
envman/
├── cmd/
│   ├── server/      # HTTP API
│   ├── migrate/     # one-shot migrator
│   └── envman/      # CLI
│       └── commands/
├── internal/
│   ├── cli/         # client-side crypto + HTTP client + config
│   ├── auth/        # register/login
│   ├── users/       # public-key lookup
│   ├── projects/    # project CRUD
│   ├── members/     # invite / wrapped project keys
│   ├── secrets/     # push/get ciphertext
│   ├── audit/       # audit log reads
│   ├── middleware/  # auth + body-size limit
│   ├── permissions/ # role check
│   ├── token/       # JWT manager (HS256 pinned)
│   ├── response/    # standardized envelopes
│   ├── config/      # env-var loading
│   ├── db/          # pgx pool
│   ├── models/      # shared DB structs
│   └── common/      # shared DB helpers
└── migrations/      # SQL + go:embed
```

---

## Roadmap

- [ ] `dev` / `staging` / `production` environments per project
- [ ] secret versioning and rollback
- [ ] `envman logs` for audit history
- [ ] Homebrew + apt packages
- [ ] email notifications on push

---

## License

MIT — see [`LICENSE`](./LICENSE).
