package projects

type ProjectRequest struct {
	Name string `json:"name"`
}

type CreateProjectResponse struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
}
