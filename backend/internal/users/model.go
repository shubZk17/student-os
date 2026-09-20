package users

import (
	"time"
)

type StudentProfile struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	FullName           string    `json:"full_name"`
	AvatarURL          string    `json:"avatar_url"`
	College            string    `json:"college"`
	Degree             string    `json:"degree"`
	Branch             string    `json:"branch"`
	CurrentYear        int       `json:"current_year"`
	GraduationYear     int       `json:"graduation_year"`
	CGPA               float64   `json:"cgpa"`
	TargetRoles        []string  `json:"target_roles"`
	PreferredLocations []string  `json:"preferred_locations"`
	WorkPreference     string    `json:"work_preference"`
	Bio                string    `json:"bio"`
	GithubURL          string    `json:"github_url"`
	LinkedinURL        string    `json:"linkedin_url"`
	PortfolioURL       string    `json:"portfolio_url"`
	ResumeURL          string    `json:"resume_url"`
	ProfileStrength    int       `json:"profile_strength"`
	Skills             []Skill   `json:"skills"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Proficiency string `json:"proficiency,omitempty"`
}

// PortfolioURL, ResumeURL and Skills are left unchanged when omitted.
type UpdateProfileRequest struct {
	FullName           string    `json:"full_name" binding:"max=255"`
	College            string    `json:"college" binding:"max=255"`
	Degree             string    `json:"degree" binding:"max=128"`
	Branch             string    `json:"branch" binding:"max=128"`
	CurrentYear        int       `json:"current_year" binding:"min=1,max=5"`
	GraduationYear     int       `json:"graduation_year" binding:"min=1950,max=2100"`
	CGPA               float64   `json:"cgpa" binding:"min=0,max=10"`
	TargetRoles        []string  `json:"target_roles" binding:"max=20,dive,max=128"`
	PreferredLocations []string  `json:"preferred_locations" binding:"max=20,dive,max=128"`
	WorkPreference     string    `json:"work_preference" binding:"omitempty,oneof=ANY REMOTE ONSITE HYBRID"`
	Bio                string    `json:"bio" binding:"max=5000"`
	GithubURL          string    `json:"github_url" binding:"omitempty,http_url,max=2048"`
	LinkedinURL        string    `json:"linkedin_url" binding:"omitempty,http_url,max=2048"`
	PortfolioURL       *string   `json:"portfolio_url" binding:"omitempty,max=2048"`
	ResumeURL          *string   `json:"resume_url" binding:"omitempty,max=2048"`
	Skills             *[]string `json:"skills" binding:"omitempty,max=50,dive,required,max=100"`
}
