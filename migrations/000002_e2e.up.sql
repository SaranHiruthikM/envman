-- Adds client-side E2E encryption primitives.
-- Each user has a long-term X25519 keypair; the private half is stored
-- encrypted-with-a-password-derived key (argon2id + XChaCha20-Poly1305).
-- Each project has a random symmetric key, wrapped per-member via ECDH.

ALTER TABLE users
    ADD COLUMN public_key                    BYTEA NOT NULL,
    ADD COLUMN encrypted_private_key         BYTEA NOT NULL,
    ADD COLUMN encrypted_private_key_nonce   BYTEA NOT NULL,
    ADD COLUMN kdf_salt                      BYTEA NOT NULL,
    ADD COLUMN kdf_time                      INTEGER NOT NULL,
    ADD COLUMN kdf_memory                    INTEGER NOT NULL,
    ADD COLUMN kdf_parallelism               INTEGER NOT NULL;

CREATE TABLE IF NOT EXISTS project_keys (
    project_id         UUID      NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id            UUID      NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    wrapped_key        BYTEA     NOT NULL,
    wrapped_key_nonce  BYTEA     NOT NULL,
    ephemeral_pubkey   BYTEA     NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id)
);
