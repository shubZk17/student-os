package jobs

import (
	"time"

	"studentos/backend/internal/matching"
)

type Opportunity struct {
	ID                string     `json:"id"`
	ExternalID        string     `json:"external_id,omitempty"`
	Source            string     `json:"source"`
	SourceURL         string     `json:"source_url"`
	Type              string     `json:"type"` // 'JOB', 'INTERNSHIP', 'HACKATHON', 'RESEARCH', 'SCHOLARSHIP'
	CompanyName       string     `json:"company_name"`
	CompanyLogoURL    string     `json:"company_logo_url"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	Location          string     `json:"location"`
	IsRemote          bool       `json:"is_remote"`
	JobType           string     `json:"job_type"` // 'INTERNSHIP', 'FULL_TIME', 'CONTRACT'
	ExperienceLevel   string     `json:"experience_level"`
	RequiredSkills    []string   `json:"required_skills"`
	EligibleDegrees   []string   `json:"eligible_degrees"`
	EligibleGradYears []int      `json:"eligible_grad_years"`
	MinCGPA           float64    `json:"min_cgpa"`
	StipendOrSalary   string     `json:"stipend_or_salary"`
	Deadline          *time.Time `json:"deadline"`
	PostedAt          time.Time  `json:"posted_at"`
	Status            string     `json:"status"`

	// Enriched fields for recommendations
	MatchScore          int      `json:"match_score,omitempty"`
	MatchedReasons      []string `json:"matched_reasons,omitempty"`
	MissingRequirements []string `json:"missing_requirements,omitempty"`
	IsSaved             bool     `json:"is_saved"`
	ApplicationStage    string   `json:"application_stage,omitempty"`
}

type RecommendationResponse struct {
	Total           int                    `json:"total"`
	Recommendations []matching.MatchResult `json:"recommendations"`
	Items           []Opportunity          `json:"items"`
}
