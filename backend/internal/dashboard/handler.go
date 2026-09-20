package dashboard

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"studentos/backend/internal/database"
	"studentos/backend/internal/middleware"
)

type SummaryResponse struct {
	StudentName             string         `json:"student_name"`
	ProfileStrength         int            `json:"profile_strength"`
	ActiveApplicationsCount int            `json:"active_applications_count"`
	UpcomingInterviewsCount int            `json:"upcoming_interviews_count"`
	ImpendingDeadlinesCount int            `json:"impending_deadlines_count"`
	UpcomingInterview       *InterviewInfo `json:"upcoming_interview,omitempty"`
}

type InterviewInfo struct {
	Company string    `json:"company"`
	Title   string    `json:"title"`
	Date    time.Time `json:"date"`
}

type Handler struct {
	db *database.DB
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) GetSummary(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var resp SummaryResponse
	err := h.db.Pool.QueryRow(ctx, `
		SELECT sp.full_name, sp.profile_strength,
		       (SELECT COUNT(*) FROM applications a
		         WHERE a.user_id = sp.user_id AND a.stage IN ('APPLIED', 'ASSESSMENT', 'INTERVIEW')),
		       (SELECT COUNT(*) FROM applications a
		         WHERE a.user_id = sp.user_id AND a.interview_date > NOW()),
		       (SELECT COUNT(*) FROM applications a JOIN opportunities o ON o.id = a.opportunity_id
		         WHERE a.user_id = sp.user_id AND a.stage = 'SAVED'
		           AND o.deadline BETWEEN NOW() AND NOW() + INTERVAL '7 days')
		FROM student_profiles sp WHERE sp.user_id = $1`, userID,
	).Scan(&resp.StudentName, &resp.ProfileStrength,
		&resp.ActiveApplicationsCount, &resp.UpcomingInterviewsCount, &resp.ImpendingDeadlinesCount)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	if err != nil {
		log.Printf("[ERROR] dashboard summary: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dashboard"})
		return
	}

	var next InterviewInfo
	err = h.db.Pool.QueryRow(ctx, `
		SELECT o.company_name, o.title, a.interview_date
		FROM applications a JOIN opportunities o ON o.id = a.opportunity_id
		WHERE a.user_id = $1 AND a.interview_date > NOW()
		ORDER BY a.interview_date LIMIT 1`, userID,
	).Scan(&next.Company, &next.Title, &next.Date)
	switch {
	case err == nil:
		resp.UpcomingInterview = &next
	case !errors.Is(err, pgx.ErrNoRows):
		log.Printf("[ERROR] dashboard next interview: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dashboard"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
