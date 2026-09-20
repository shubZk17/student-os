package jobs

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"studentos/backend/internal/database"
	"studentos/backend/internal/matching"
	"studentos/backend/internal/middleware"
)

type Handler struct {
	db     *database.DB
	engine *matching.Engine
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{
		db:     db,
		engine: matching.NewEngine(matching.DefaultWeights),
	}
}

const opportunityColumns = `
	o.id, COALESCE(o.external_id, ''), o.source, o.source_url, o.type, o.company_name, COALESCE(o.company_logo_url, ''),
	o.title, o.description, o.location, o.is_remote, COALESCE(o.job_type, ''), COALESCE(o.experience_level, ''),
	COALESCE((SELECT array_agg(s.name) FROM opportunity_skills os JOIN skills s ON s.id = os.skill_id
	          WHERE os.opportunity_id = o.id AND os.is_required), '{}'),
	COALESCE(o.eligible_degrees, '{}'), COALESCE(o.eligible_grad_years, '{}'), COALESCE(o.min_cgpa, 0)::float8,
	COALESCE(o.stipend_or_salary, ''), o.deadline, o.posted_at, o.status`

func scanOpportunity(row pgx.Row, o *Opportunity) error {
	return row.Scan(
		&o.ID, &o.ExternalID, &o.Source, &o.SourceURL, &o.Type, &o.CompanyName, &o.CompanyLogoURL,
		&o.Title, &o.Description, &o.Location, &o.IsRemote, &o.JobType, &o.ExperienceLevel,
		&o.RequiredSkills, &o.EligibleDegrees, &o.EligibleGradYears, &o.MinCGPA,
		&o.StipendOrSalary, &o.Deadline, &o.PostedAt, &o.Status,
	)
}

// GetRecommendations computes personalized, explainable job matches for the authenticated student
func (h *Handler) GetRecommendations(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	candidate, err := h.getCandidateProfile(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	if err != nil {
		log.Printf("[ERROR] recommendations: load profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load profile"})
		return
	}

	// ponytail: scores every active opportunity in memory per request; pre-filter in SQL
	// (degree/grad year/deadline) or cache results once the catalog reaches thousands of rows.
	rows, err := h.db.Pool.Query(ctx, `SELECT `+opportunityColumns+` FROM opportunities o
		WHERE o.status = 'ACTIVE' AND (o.deadline IS NULL OR o.deadline > NOW())`)
	if err != nil {
		log.Printf("[ERROR] recommendations: query opportunities: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load opportunities"})
		return
	}
	defer rows.Close()

	matched := make([]Opportunity, 0)
	for rows.Next() {
		var opp Opportunity
		if err := scanOpportunity(rows, &opp); err != nil {
			log.Printf("[ERROR] recommendations: scan: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load opportunities"})
			return
		}
		res, eligible := h.engine.Evaluate(candidate, &matching.OpportunityItem{
			ID:                opp.ID,
			Title:             opp.Title,
			CompanyName:       opp.CompanyName,
			Description:       opp.Description,
			Location:          opp.Location,
			IsRemote:          opp.IsRemote,
			RequiredSkills:    opp.RequiredSkills,
			EligibleDegrees:   opp.EligibleDegrees,
			EligibleGradYears: opp.EligibleGradYears,
			MinCGPA:           opp.MinCGPA,
			Deadline:          opp.Deadline,
			Status:            opp.Status,
		}, 0.0)
		if eligible {
			opp.MatchScore = res.MatchPercentage
			opp.MatchedReasons = res.MatchedReasons
			opp.MissingRequirements = res.MissingRequirements
			matched = append(matched, opp)
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] recommendations: rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load opportunities"})
		return
	}

	sort.SliceStable(matched, func(i, j int) bool {
		return matched[i].MatchScore > matched[j].MatchScore
	})

	c.JSON(http.StatusOK, gin.H{
		"total":           len(matched),
		"recommendations": matched,
	})
}

// ListOpportunities provides searchable, filtered, paginated list of active opportunities
func (h *Handler) ListOpportunities(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Empty filter params are no-ops.
	query := `SELECT ` + opportunityColumns + `, COUNT(*) OVER()
		FROM opportunities o
		WHERE o.status = 'ACTIVE'
		  AND ($1 = '' OR upper(o.type) = upper($1))
		  AND (NOT $2 OR o.is_remote)
		  AND ($3 = '' OR position(lower($3) IN lower(o.location)) > 0)
		  AND ($4 = '' OR position(lower($4) IN lower(o.title || ' ' || o.company_name || ' ' || o.description)) > 0)
		ORDER BY o.posted_at DESC
		LIMIT $5 OFFSET $6`
	rows, err := h.db.Pool.Query(ctx, query,
		strings.TrimSpace(c.Query("type")),
		c.Query("remote") == "true",
		strings.TrimSpace(c.Query("location")),
		strings.TrimSpace(c.Query("q")),
		limit, (page-1)*limit,
	)
	if err != nil {
		log.Printf("[ERROR] list opportunities: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load opportunities"})
		return
	}
	defer rows.Close()

	items := make([]Opportunity, 0)
	total := 0
	for rows.Next() {
		var o Opportunity
		if err := rows.Scan(
			&o.ID, &o.ExternalID, &o.Source, &o.SourceURL, &o.Type, &o.CompanyName, &o.CompanyLogoURL,
			&o.Title, &o.Description, &o.Location, &o.IsRemote, &o.JobType, &o.ExperienceLevel,
			&o.RequiredSkills, &o.EligibleDegrees, &o.EligibleGradYears, &o.MinCGPA,
			&o.StipendOrSalary, &o.Deadline, &o.PostedAt, &o.Status, &total,
		); err != nil {
			log.Printf("[ERROR] list opportunities: scan: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load opportunities"})
			return
		}
		items = append(items, o)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] list opportunities: rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load opportunities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":  page,
		"limit": limit,
		"total": total,
		"items": items,
	})
}

// GetOpportunityByID returns details of a single opportunity
func (h *Handler) GetOpportunityByID(c *gin.Context) {
	id := c.Param("id")
	if uuid.Validate(id) != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Opportunity not found"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var opp Opportunity
	err := scanOpportunity(h.db.Pool.QueryRow(ctx, `SELECT `+opportunityColumns+` FROM opportunities o WHERE o.id = $1`, id), &opp)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Opportunity not found"})
		return
	}
	if err != nil {
		log.Printf("[ERROR] get opportunity: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load opportunity"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"opportunity": opp})
}

func (h *Handler) getCandidateProfile(ctx context.Context, userID string) (*matching.CandidateProfile, error) {
	cand := &matching.CandidateProfile{ID: userID, AppliedJobIDs: make(map[string]bool)}
	var appliedIDs []string

	err := h.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(sp.degree, ''), COALESCE(sp.branch, ''), sp.graduation_year, COALESCE(sp.cgpa, 0)::float8,
		       COALESCE(sp.target_roles, '{}'), COALESCE(sp.preferred_locations, '{}'), COALESCE(sp.work_preference, 'ANY'),
		       COALESCE((SELECT array_agg(s.name) FROM student_skills ss JOIN skills s ON s.id = ss.skill_id
		                 WHERE ss.student_id = sp.id), '{}'),
		       COALESCE((SELECT array_agg(DISTINCT s.name) FROM projects p
		                 JOIN project_skills ps ON ps.project_id = p.id JOIN skills s ON s.id = ps.skill_id
		                 WHERE p.student_id = sp.id), '{}'),
		       COALESCE((SELECT array_agg(a.opportunity_id::text) FROM applications a
		                 WHERE a.user_id = sp.user_id AND a.stage <> 'SAVED'), '{}')
		FROM student_profiles sp WHERE sp.user_id = $1`, userID,
	).Scan(&cand.Degree, &cand.Branch, &cand.GraduationYear, &cand.CGPA,
		&cand.TargetRoles, &cand.PreferredLocations, &cand.WorkPreference,
		&cand.Skills, &cand.ProjectSkills, &appliedIDs)
	if err != nil {
		return nil, err
	}
	for _, id := range appliedIDs {
		cand.AppliedJobIDs[id] = true
	}
	return cand, nil
}
