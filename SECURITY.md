# Security policy

envman stores encrypted secrets for its users, so vulnerabilities in it have
meaningful blast radius. We take reports seriously and will work with you to
fix them quickly and publicly credit you if you want.

## Reporting a vulnerability

**Please do not file a public GitHub issue.** Instead, email the maintainer:

- **security@envman** *(replace with the real monitored address before launch)*

Include:

- a description of the issue and its impact
- steps to reproduce (or a proof-of-concept)
- the commit, release, or deployed version you tested against
- your name/handle if you'd like credit (optional)

You can expect an acknowledgement within **72 hours** and a first substantive
response within **7 days**. Issues will be fixed, released, and disclosed within
**90 days** of the first report, or sooner for critical findings.

PGP: *(publish a key fingerprint here before launch)*

## Scope

In scope for reports:

- the envman CLI (`cmd/envman/`)
- the envman server (`cmd/server/`) and its HTTP API
- the encryption model (`internal/cli/keys.go`, `internal/cli/crypto.go`)
- the database schema (`migrations/`) and migration flow

Out of scope:

- vulnerabilities in third-party dependencies that are already publicly known
  — please report those upstream and mention them to us if we need to upgrade
- social engineering of the project maintainers
- denial-of-service against our hosted instance from a single IP at normal
  rate limits (we handle that at the edge)
- missing security headers on `/healthz` and `/readyz`

## What we commit to

- **Encryption model**: secrets are encrypted on the client with a per-project
  symmetric key, which itself is wrapped per-member via X25519 ECDH + HKDF +
  XChaCha20-Poly1305. A full database dump reveals only ciphertext and wrapped
  keys — no plaintext, no decryption keys.
- **Password handling**: passwords are bcrypt-hashed server-side; the KDF for
  the private-key wrapper is argon2id and runs client-side only.
- **Transport**: HTTPS in production, enforced by the deployment platform.
- **Audit**: every push is logged with the pushing user and a timestamp.

## What we do not yet commit to

- a formal bug bounty (watch this file — may change at launch)
- 24/7 response — we are a small team

Thank you for helping keep envman safe.
