package auth

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`

	PublicKey                string `json:"public_key"`
	EncryptedPrivateKey      string `json:"encrypted_private_key"`
	EncryptedPrivateKeyNonce string `json:"encrypted_private_key_nonce"`
	KDFSalt                  string `json:"kdf_salt"`
	KDFTime                  uint32 `json:"kdf_time"`
	KDFMemory                uint32 `json:"kdf_memory"`
	KDFParallelism           uint8  `json:"kdf_parallelism"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

type LoginResponse struct {
	Token                    string `json:"token"`
	UserID                   string `json:"user_id"`
	PublicKey                string `json:"public_key"`
	EncryptedPrivateKey      string `json:"encrypted_private_key"`
	EncryptedPrivateKeyNonce string `json:"encrypted_private_key_nonce"`
	KDFSalt                  string `json:"kdf_salt"`
	KDFTime                  uint32 `json:"kdf_time"`
	KDFMemory                uint32 `json:"kdf_memory"`
	KDFParallelism           uint8  `json:"kdf_parallelism"`
}
