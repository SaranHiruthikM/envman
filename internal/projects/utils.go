package projects

import "github.com/SaranHiruthikM/envman/internal/models"

type ProjectRequest struct {
	Name string `json:"name"`
}

type CreateProjectResponse struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
}

type GetProjectResponse struct {
	Project *models.Project `json:"project"`
}
