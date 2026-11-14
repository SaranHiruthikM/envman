package cli

import (
	"crypto/ecdh"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
	"crypto/sha256"
)

// Argon2idParams are tuned for interactive logins: ~250ms on a modern laptop,
// hard enough that a stolen database row costs real money to crack.
type Argon2idParams struct {
	Time        uint32 `json:"time"`
	Memory      uint32 `json:"memory"`
	Parallelism uint8  `json:"parallelism"`
}

func DefaultArgon2idParams() Argon2idParams {
	return Argon2idParams{Time: 3, Memory: 64 * 1024, Parallelism: 2}
}

const (
	kekLen            = 32
	projectKeyLen     = 32
	kdfSaltLen        = 16
	xNonceLen         = chacha20poly1305.NonceSizeX // 24
)

// DeriveKEK turns a password + salt into a 32-byte key-encryption-key.
func DeriveKEK(password string, salt []byte, p Argon2idParams) []byte {
	return argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Parallelism, kekLen)
}

func RandomSalt() ([]byte, error) {
	s := make([]byte, kdfSaltLen)
	if _, err := io.ReadFull(rand.Reader, s); err != nil {
		return nil, err
	}
	return s, nil
}

func RandomProjectKey() ([]byte, error) {
	k := make([]byte, projectKeyLen)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil, err
	}
	return k, nil
}

// GenerateUserKeypair returns (priv, pub) on the X25519 curve.
func GenerateUserKeypair() (priv, pub []byte, err error) {
	curve := ecdh.X25519()
	privKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return privKey.Bytes(), privKey.PublicKey().Bytes(), nil
}

// EncryptPrivateKey wraps the user's private key with the password-derived KEK.
func EncryptPrivateKey(priv, kek []byte) (ciphertext, nonce []byte, err error) {
	aead, err := chacha20poly1305.NewX(kek)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, xNonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	return aead.Seal(nil, nonce, priv, nil), nonce, nil
}

func DecryptPrivateKey(ciphertext, nonce, kek []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(kek)
	if err != nil {
		return nil, err
	}
	return aead.Open(nil, nonce, ciphertext, nil)
}

// WrapProjectKey produces an ephemeral X25519 key, performs ECDH with the
// recipient's public key, derives a wrapping key via HKDF-SHA256, then AEAD-wraps
// the project key. Returns (ephemeralPub, nonce, wrappedKey).
func WrapProjectKey(projectKey, recipientPub []byte) (ephPub, nonce, wrapped []byte, err error) {
	curve := ecdh.X25519()

	recipient, err := curve.NewPublicKey(recipientPub)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid recipient public key: %w", err)
	}
	eph, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}
	shared, err := eph.ECDH(recipient)
	if err != nil {
		return nil, nil, nil, err
	}

	wrapKey, err := hkdfExpand(shared, []byte("envman/project-key-wrap/v1"), 32)
	if err != nil {
		return nil, nil, nil, err
	}

	aead, err := chacha20poly1305.NewX(wrapKey)
	if err != nil {
		return nil, nil, nil, err
	}
	nonce = make([]byte, xNonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, nil, err
	}
	wrapped = aead.Seal(nil, nonce, projectKey, nil)

	return eph.PublicKey().Bytes(), nonce, wrapped, nil
}

// UnwrapProjectKey reverses WrapProjectKey using the recipient's own private key.
func UnwrapProjectKey(ephPub, nonce, wrapped, myPriv []byte) ([]byte, error) {
	curve := ecdh.X25519()

	myKey, err := curve.NewPrivateKey(myPriv)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	peerPub, err := curve.NewPublicKey(ephPub)
	if err != nil {
		return nil, fmt.Errorf("invalid ephemeral public key: %w", err)
	}
	shared, err := myKey.ECDH(peerPub)
	if err != nil {
		return nil, err
	}

	wrapKey, err := hkdfExpand(shared, []byte("envman/project-key-wrap/v1"), 32)
	if err != nil {
		return nil, err
	}

	aead, err := chacha20poly1305.NewX(wrapKey)
	if err != nil {
		return nil, err
	}
	return aead.Open(nil, nonce, wrapped, nil)
}

func hkdfExpand(secret, info []byte, size int) ([]byte, error) {
	r := hkdf.New(sha256.New, secret, nil, info)
	out := make([]byte, size)
	if _, err := io.ReadFull(r, out); err != nil {
		return nil, err
	}
	return out, nil
}
