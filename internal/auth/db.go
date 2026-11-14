package auth

import "context"

type createUserParams struct {
	Name, Email, PasswordHash string

	PublicKey                []byte
	EncryptedPrivateKey      []byte
	EncryptedPrivateKeyNonce []byte
	KDFSalt                  []byte
	KDFTime                  uint32
	KDFMemory                uint32
	KDFParallelism           uint8
}

func (h *Handler) createUser(ctx context.Context, p createUserParams) (string, error) {
	const query = `
		INSERT INTO users (
			id, email, name, password_hash, created_at,
			public_key, encrypted_private_key, encrypted_private_key_nonce,
			kdf_salt, kdf_time, kdf_memory, kdf_parallelism
		) VALUES (
			gen_random_uuid(), $1, $2, $3, NOW(),
			$4, $5, $6, $7, $8, $9, $10
		) RETURNING id`

	var id string
	err := h.db.QueryRow(ctx, query,
		p.Email, p.Name, p.PasswordHash,
		p.PublicKey, p.EncryptedPrivateKey, p.EncryptedPrivateKeyNonce,
		p.KDFSalt, p.KDFTime, p.KDFMemory, p.KDFParallelism,
	).Scan(&id)
	return id, err
}
