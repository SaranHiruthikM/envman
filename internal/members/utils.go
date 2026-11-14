package members

type InviteMembersRequest struct {
	Email           string `json:"email"`
	Role            string `json:"role"`
	WrappedKey      string `json:"wrapped_key"`
	WrappedKeyNonce string `json:"wrapped_key_nonce"`
	EphemeralPubkey string `json:"ephemeral_pubkey"`
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}

type SetSelfKeyRequest struct {
	WrappedKey      string `json:"wrapped_key"`
	WrappedKeyNonce string `json:"wrapped_key_nonce"`
	EphemeralPubkey string `json:"ephemeral_pubkey"`
}
