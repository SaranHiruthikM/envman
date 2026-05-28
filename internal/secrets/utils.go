package secrets

type PushSecretsRequest struct {
	EncryptedBlob string `json:"encryptedData"`
}
