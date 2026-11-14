package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`

	// E2E key material. Public key is shareable; everything else is
	// opaque to the server (stored, never inspected).
	PublicKey                []byte `json:"-"`
	EncryptedPrivateKey      []byte `json:"-"`
	EncryptedPrivateKeyNonce []byte `json:"-"`
	KDFSalt                  []byte `json:"-"`
	KDFTime                  uint32 `json:"-"`
	KDFMemory                uint32 `json:"-"`
	KDFParallelism           uint8  `json:"-"`
}

type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedBy *string   `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Logs struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	PushedBy  *string   `json:"pushed_by,omitempty"`
	SecretsID *string   `json:"secrets_id,omitempty"`
	PushedAt  time.Time `json:"pushed_at"`
}
