package projects

import "time"

type Project struct {
	ID           string    `json:"id"`
	StudentID    string    `json:"student_id"`
	Title        string    `json:"title"`
	Tagline      string    `json:"tagline"`
	Description  string    `json:"description"`
	Role         string    `json:"role"`
	GithubURL    string    `json:"github_url"`
	DemoURL      string    `json:"demo_url"`
	Metrics      string    `json:"metrics"`
	Technologies []string  `json:"technologies"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateProjectRequest struct {
	Title        string   `json:"title" binding:"required,max=255"`
	Tagline      string   `json:"tagline" binding:"max=255"`
	Description  string   `json:"description" binding:"required,max=10000"`
	Role         string   `json:"role" binding:"max=128"`
	GithubURL    string   `json:"github_url" binding:"omitempty,http_url,max=2048"`
	DemoURL      string   `json:"demo_url" binding:"omitempty,http_url,max=2048"`
	Metrics      string   `json:"metrics" binding:"max=2000"`
	Technologies []string `json:"technologies" binding:"max=30,dive,required,max=100"`
}
