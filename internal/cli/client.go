package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	apiURL string
	http   *http.Client
}

func NewClient(apiURL string) *Client {
	return &Client{
		apiURL: apiURL,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

// ErrSessionExpired is returned on 401 so callers can prompt re-login.
var ErrSessionExpired = errors.New("session expired — run `envman login`")

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) do(method, path string, body any, token string, out any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.apiURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusUnauthorized {
			return ErrSessionExpired
		}
		var apiErr apiError
		if json.Unmarshal(raw, &apiErr) == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("%s (%d)", apiErr.Error.Message, resp.StatusCode)
		}
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(raw))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("malformed response: %w", err)
		}
	}
	return nil
}

// ---- auth ----

type RegisterRequest struct {
	Email                    string `json:"email"`
	Name                     string `json:"name"`
	Password                 string `json:"password"`
	PublicKey                string `json:"public_key"`
	EncryptedPrivateKey      string `json:"encrypted_private_key"`
	EncryptedPrivateKeyNonce string `json:"encrypted_private_key_nonce"`
	KDFSalt                  string `json:"kdf_salt"`
	KDFTime                  uint32 `json:"kdf_time"`
	KDFMemory                uint32 `json:"kdf_memory"`
	KDFParallelism           uint8  `json:"kdf_parallelism"`
}

type AuthResponse struct {
	Data struct {
		Token  string `json:"token"`
		UserID string `json:"user_id"`
	} `json:"data"`
}

type LoginResponse struct {
	Data struct {
		Token                    string `json:"token"`
		UserID                   string `json:"user_id"`
		PublicKey                string `json:"public_key"`
		EncryptedPrivateKey      string `json:"encrypted_private_key"`
		EncryptedPrivateKeyNonce string `json:"encrypted_private_key_nonce"`
		KDFSalt                  string `json:"kdf_salt"`
		KDFTime                  uint32 `json:"kdf_time"`
		KDFMemory                uint32 `json:"kdf_memory"`
		KDFParallelism           uint8  `json:"kdf_parallelism"`
	} `json:"data"`
}

func (c *Client) Register(req RegisterRequest) (*AuthResponse, error) {
	var out AuthResponse
	if err := c.do("POST", "/auth/register", req, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Login(email, password string) (*LoginResponse, error) {
	var out LoginResponse
	err := c.do("POST", "/auth/login",
		map[string]string{"email": email, "password": password}, "", &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ---- projects ----

type CreateProjectResponse struct {
	Data struct {
		ProjectID string `json:"project_id"`
		Name      string `json:"name"`
	} `json:"data"`
}

func (c *Client) CreateProject(token, name string) (string, error) {
	var out CreateProjectResponse
	err := c.do("POST", "/projects",
		map[string]string{"name": name}, token, &out)
	return out.Data.ProjectID, err
}

// ---- public-key lookup for invite ----

type PublicKeyResponse struct {
	Data struct {
		UserID    string `json:"user_id"`
		PublicKey string `json:"public_key"`
	} `json:"data"`
}

func (c *Client) GetUserPublicKey(token, email string) (string, string, error) {
	var out PublicKeyResponse
	err := c.do("GET", "/users/by-email/"+email+"/public-key", nil, token, &out)
	return out.Data.UserID, out.Data.PublicKey, err
}

// ---- project keys ----

type ProjectKeyPayload struct {
	WrappedKey      string `json:"wrapped_key"`
	WrappedKeyNonce string `json:"wrapped_key_nonce"`
	EphemeralPubkey string `json:"ephemeral_pubkey"`
}

type ProjectKeyResponse struct {
	Data ProjectKeyPayload `json:"data"`
}

func (c *Client) SetSelfKey(token, projectID string, p ProjectKeyPayload) error {
	return c.do("POST", "/projects/"+projectID+"/keys", p, token, nil)
}

func (c *Client) GetSelfKey(token, projectID string) (*ProjectKeyPayload, error) {
	var out ProjectKeyResponse
	err := c.do("GET", "/projects/"+projectID+"/keys/me", nil, token, &out)
	if err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// ---- members ----

type InviteMemberRequest struct {
	Email           string `json:"email"`
	Role            string `json:"role"`
	WrappedKey      string `json:"wrapped_key"`
	WrappedKeyNonce string `json:"wrapped_key_nonce"`
	EphemeralPubkey string `json:"ephemeral_pubkey"`
}

func (c *Client) InviteMember(token, projectID string, req InviteMemberRequest) error {
	return c.do("POST", "/projects/"+projectID+"/members", req, token, nil)
}

func (c *Client) UpdateRole(token, projectID, userID, role string) error {
	return c.do("PATCH", "/projects/"+projectID+"/members/"+userID,
		map[string]string{"role": role}, token, nil)
}

func (c *Client) RemoveMember(token, projectID, userID string) error {
	return c.do("DELETE", "/projects/"+projectID+"/members/"+userID, nil, token, nil)
}

// ---- secrets ----

type SecretsResponse struct {
	Data struct {
		Ciphertext string `json:"ciphertext"`
	} `json:"data"`
}

func (c *Client) PushSecrets(token, projectID, ciphertext string) error {
	return c.do("POST", "/projects/"+projectID+"/secrets",
		map[string]string{"ciphertext": ciphertext}, token, nil)
}

func (c *Client) GetSecrets(token, projectID string) (string, error) {
	var out SecretsResponse
	err := c.do("GET", "/projects/"+projectID+"/secrets", nil, token, &out)
	return out.Data.Ciphertext, err
}
