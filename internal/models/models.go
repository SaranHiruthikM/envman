package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type Logs struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	PushedBy  string    `json:"pushed_by"`
	SecretsID string    `json:"secrets_id"`
	PushedAt  time.Time `json:"pushed_at"`
}
