package applications

import "time"

type Application struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	OpportunityID  string     `json:"opportunity_id"`
	Company        string     `json:"company_name"`
	Title          string     `json:"title"`
	Stage          string     `json:"stage"` // 'SAVED', 'APPLIED', 'ASSESSMENT', 'INTERVIEW', 'OFFER', 'REJECTED', 'WITHDRAWN'
	AppliedAt      *time.Time `json:"applied_at"`
	InterviewDate  *time.Time `json:"interview_date"`
	NextFollowUp   *time.Time `json:"next_follow_up"`
	Notes          string     `json:"notes"`
	Location       string     `json:"location"`
	CompanyLogoURL string     `json:"company_logo_url"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateApplicationRequest struct {
	OpportunityID string     `json:"opportunity_id" binding:"required,uuid"`
	Stage         string     `json:"stage" binding:"omitempty,oneof=SAVED APPLIED ASSESSMENT INTERVIEW OFFER REJECTED WITHDRAWN"` // default 'SAVED'
	Notes         string     `json:"notes" binding:"max=5000"`
	InterviewDate *time.Time `json:"interview_date"`
}

// Omitted optional fields leave the stored value unchanged.
type UpdateStageRequest struct {
	Stage         string     `json:"stage" binding:"required,oneof=SAVED APPLIED ASSESSMENT INTERVIEW OFFER REJECTED WITHDRAWN"`
	Notes         *string    `json:"notes" binding:"omitempty,max=5000"`
	InterviewDate *time.Time `json:"interview_date"`
	NextFollowUp  *time.Time `json:"next_follow_up"`
}
