package members

type InviteMembersRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}
