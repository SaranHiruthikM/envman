package projects

type ProjectRequest struct {
	Name string `json:"name"`
}

type ProjectResponse struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
}
