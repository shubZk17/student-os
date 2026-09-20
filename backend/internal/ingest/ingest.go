// Package ingest pulls student roles from public job boards into the opportunities
// table. It is called by cmd/worker (CLI / cron) and by the API's scheduled-ingest
// route, so the logic lives here rather than in a main package.
package ingest

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"studentos/backend/internal/database"
	"studentos/backend/internal/jobs/provider"
)

// Board lists are overridable via env: comma-separated "slug" (Greenhouse) or "slug:Company Name".
const (
	defaultGreenhouse = "stripe,airbnb,datadog,coinbase,figma,brex,gusto,anthropic,cloudflare,scaleai,vercel,razorpaysoftwareprivatelimited:Razorpay"
	defaultLever      = "spotify:Spotify"
	defaultAshby      = "notion:Notion,plaid:Plaid,openai:OpenAI,linear:Linear"
)

// Timeout bounds a full ingestion run.
const Timeout = 15 * time.Minute

// Result summarises one ingestion run.
type Result struct {
	Opportunities int `json:"opportunities"`
	BoardsOK      int `json:"boards_ok"`
	BoardsFailed  int `json:"boards_failed"`
}

// TotalOutage reports whether every board failed, which callers treat as a failed run.
func (r Result) TotalOutage() bool {
	return r.BoardsOK == 0 && r.BoardsFailed > 0
}

// Run syncs every configured board. Per-board failures are logged and skipped so one
// dead board cannot fail the whole run.
func Run(ctx context.Context, db *database.DB) (Result, error) {
	var res Result

	var known []string
	if err := db.Pool.QueryRow(ctx, `SELECT COALESCE(array_agg(name), '{}') FROM skills`).Scan(&known); err != nil {
		return res, err
	}
	skills := provider.NewSkillMatcher(known)

	boards := append(append(
		parseBoards("greenhouse", envOr("GREENHOUSE_BOARDS", defaultGreenhouse)),
		parseBoards("lever", envOr("LEVER_COMPANIES", defaultLever))...),
		parseBoards("ashby", envOr("ASHBY_BOARDS", defaultAshby))...)

	client := &http.Client{Timeout: 60 * time.Second}
	for _, b := range boards {
		n, err := syncBoard(ctx, db, client, skills, b)
		if err != nil {
			res.BoardsFailed++
			log.Printf("[ERROR] %s/%s: %v", b.Source, b.Slug, err)
			continue
		}
		res.BoardsOK++
		res.Opportunities += n
		log.Printf("[INGESTION] %s/%s: %d student opportunities", b.Source, b.Slug, n)
	}

	if _, err := db.Pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW() - INTERVAL '1 day'`); err != nil {
		log.Printf("[WARN] Failed to prune expired refresh tokens: %v", err)
	}

	log.Printf("[INFO] Ingestion finished: %d opportunities, %d/%d boards failed",
		res.Opportunities, res.BoardsFailed, len(boards))
	return res, nil
}

// syncBoard upserts the board's student postings and closes ones no longer listed.
// Nothing is closed if the fetch fails, so an API outage can't wipe a company's listings.
func syncBoard(ctx context.Context, db *database.DB, client *http.Client, skills *provider.SkillMatcher, b provider.Board) (int, error) {
	jobs, err := provider.Fetch(ctx, client, b)
	if err != nil {
		return 0, err
	}

	seen := []string{}
	for _, j := range jobs {
		if !provider.IsStudentRole(j.Title) {
			continue
		}
		j.RequiredSkills = skills.Match(j.Title + "\n" + j.Description)
		if err := upsertJob(ctx, db, j); err != nil {
			log.Printf("[WARN] Failed to upsert job %s: %v", j.ExternalID, err)
			continue
		}
		seen = append(seen, j.ExternalID)
	}

	_, err = db.Pool.Exec(ctx, `
		UPDATE opportunities SET status = 'CLOSED', updated_at = NOW()
		WHERE source = $1 AND starts_with(external_id, $2) AND status = 'ACTIVE' AND NOT (external_id = ANY($3))`,
		b.Source, b.ExternalIDPrefix(), seen)
	return len(seen), err
}

func upsertJob(ctx context.Context, db *database.DB, j provider.IngestedJob) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var id string
	err = tx.QueryRow(ctx, `
		INSERT INTO opportunities (
			external_id, source, source_url, type, company_name, company_logo_url,
			title, description, location, is_remote, job_type, experience_level,
			eligible_degrees, eligible_grad_years, min_cgpa, stipend_or_salary,
			deadline, posted_at, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, 'ACTIVE')
		ON CONFLICT (source, external_id) DO UPDATE SET
			source_url = EXCLUDED.source_url,
			type = EXCLUDED.type,
			company_name = EXCLUDED.company_name,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			location = EXCLUDED.location,
			is_remote = EXCLUDED.is_remote,
			job_type = EXCLUDED.job_type,
			experience_level = EXCLUDED.experience_level,
			deadline = EXCLUDED.deadline,
			status = 'ACTIVE',
			updated_at = NOW()
		RETURNING id`,
		j.ExternalID, j.Source, j.SourceURL, j.Type, j.CompanyName, j.CompanyLogoURL,
		j.Title, j.Description, j.Location, j.IsRemote, j.JobType, j.ExperienceLevel,
		j.EligibleDegrees, j.EligibleGradYears, j.MinCGPA, j.StipendOrSalary,
		j.Deadline, j.PostedAt,
	).Scan(&id)
	if err != nil {
		return err
	}

	// Skills only come from the known list, so no new skills rows are created here.
	if _, err := tx.Exec(ctx, `DELETE FROM opportunity_skills WHERE opportunity_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO opportunity_skills (opportunity_id, skill_id)
		SELECT $1, id FROM skills WHERE name = ANY($2)`, id, j.RequiredSkills); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func parseBoards(source, csv string) []provider.Board {
	var boards []provider.Board
	for _, entry := range strings.Split(csv, ",") {
		slug, company, _ := strings.Cut(strings.TrimSpace(entry), ":")
		if slug != "" {
			boards = append(boards, provider.Board{Source: source, Slug: slug, Company: strings.TrimSpace(company)})
		}
	}
	return boards
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
