// Package provider fetches public job postings from applicant tracking systems
// (Greenhouse, Lever, Ashby) and normalizes them into IngestedJob.
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type IngestedJob struct {
	ExternalID        string
	Source            string
	SourceURL         string
	Type              string // 'JOB', 'INTERNSHIP'
	CompanyName       string
	CompanyLogoURL    string
	Title             string
	Description       string
	Location          string
	IsRemote          bool
	JobType           string // 'INTERNSHIP', 'FULL_TIME'
	ExperienceLevel   string // 'INTERN', 'ENTRY'
	RequiredSkills    []string
	EligibleDegrees   []string
	EligibleGradYears []int
	MinCGPA           float64
	StipendOrSalary   string
	Deadline          *time.Time
	PostedAt          time.Time
}

// Board is one company's job board. Company is required for Lever and Ashby,
// whose APIs don't return the company name.
type Board struct {
	Source  string // "greenhouse", "lever", "ashby"
	Slug    string
	Company string
}

// ExternalIDPrefix scopes a board's postings, so the worker can close ones that disappeared.
func (b Board) ExternalIDPrefix() string { return b.Slug + ":" }

// Overridden in tests.
var (
	greenhouseURL = "https://boards-api.greenhouse.io/v1/boards/%s/jobs?content=true"
	leverURL      = "https://api.lever.co/v0/postings/%s?mode=json"
	ashbyURL      = "https://api.ashbyhq.com/posting-api/job-board/%s"
)

// Fetch returns every posting on the board, normalized. Filtering happens in the caller.
func Fetch(ctx context.Context, client *http.Client, b Board) ([]IngestedJob, error) {
	switch b.Source {
	case "greenhouse":
		return fetchGreenhouse(ctx, client, b)
	case "lever":
		return fetchLever(ctx, client, b)
	case "ashby":
		return fetchAshby(ctx, client, b)
	}
	return nil, fmt.Errorf("unknown source %q", b.Source)
}

func getJSON(ctx context.Context, client *http.Client, rawURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "StudentOS-JobSync/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", rawURL, resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 64<<20)).Decode(out)
}

func fetchGreenhouse(ctx context.Context, client *http.Client, b Board) ([]IngestedJob, error) {
	var resp struct {
		Jobs []struct {
			ID                  int64  `json:"id"`
			Title               string `json:"title"`
			AbsoluteURL         string `json:"absolute_url"`
			CompanyName         string `json:"company_name"`
			Content             string `json:"content"`
			FirstPublished      string `json:"first_published"`
			UpdatedAt           string `json:"updated_at"`
			ApplicationDeadline string `json:"application_deadline"`
			Location            struct {
				Name string `json:"name"`
			} `json:"location"`
		} `json:"jobs"`
	}
	if err := getJSON(ctx, client, fmt.Sprintf(greenhouseURL, url.PathEscape(b.Slug)), &resp); err != nil {
		return nil, err
	}
	jobs := make([]IngestedJob, 0, len(resp.Jobs))
	for _, j := range resp.Jobs {
		company := b.Company
		if company == "" {
			company = j.CompanyName
		}
		posted := parseTime(j.FirstPublished)
		if posted == nil {
			posted = parseTime(j.UpdatedAt)
		}
		jobs = append(jobs, newJob(b, fmt.Sprint(j.ID), j.AbsoluteURL, company, j.Title,
			// Greenhouse returns entity-escaped HTML.
			htmlToText(html.UnescapeString(j.Content)), j.Location.Name,
			strings.Contains(strings.ToLower(j.Location.Name), "remote"),
			false, posted, parseTime(j.ApplicationDeadline)))
	}
	return jobs, nil
}

func fetchLever(ctx context.Context, client *http.Client, b Board) ([]IngestedJob, error) {
	var resp []struct {
		ID               string `json:"id"`
		Text             string `json:"text"`
		HostedURL        string `json:"hostedUrl"`
		DescriptionPlain string `json:"descriptionPlain"`
		CreatedAt        int64  `json:"createdAt"` // unix ms
		WorkplaceType    string `json:"workplaceType"`
		Categories       struct {
			Location   string `json:"location"`
			Commitment string `json:"commitment"`
		} `json:"categories"`
	}
	if err := getJSON(ctx, client, fmt.Sprintf(leverURL, url.PathEscape(b.Slug)), &resp); err != nil {
		return nil, err
	}
	jobs := make([]IngestedJob, 0, len(resp))
	for _, j := range resp {
		posted := time.UnixMilli(j.CreatedAt).UTC()
		remote := strings.EqualFold(j.WorkplaceType, "remote") || strings.Contains(strings.ToLower(j.Categories.Location), "remote")
		jobs = append(jobs, newJob(b, j.ID, j.HostedURL, b.Company, j.Text, j.DescriptionPlain,
			j.Categories.Location, remote, strings.EqualFold(j.Categories.Commitment, "intern"), &posted, nil))
	}
	return jobs, nil
}

func fetchAshby(ctx context.Context, client *http.Client, b Board) ([]IngestedJob, error) {
	var resp struct {
		Jobs []struct {
			ID               string `json:"id"`
			Title            string `json:"title"`
			JobURL           string `json:"jobUrl"`
			DescriptionPlain string `json:"descriptionPlain"`
			PublishedAt      string `json:"publishedAt"`
			Location         string `json:"location"`
			IsRemote         bool   `json:"isRemote"`
			IsListed         bool   `json:"isListed"`
			EmploymentType   string `json:"employmentType"`
		} `json:"jobs"`
	}
	if err := getJSON(ctx, client, fmt.Sprintf(ashbyURL, url.PathEscape(b.Slug)), &resp); err != nil {
		return nil, err
	}
	jobs := make([]IngestedJob, 0, len(resp.Jobs))
	for _, j := range resp.Jobs {
		if !j.IsListed {
			continue
		}
		jobs = append(jobs, newJob(b, j.ID, j.JobURL, b.Company, j.Title, j.DescriptionPlain,
			j.Location, j.IsRemote, strings.EqualFold(j.EmploymentType, "intern"), parseTime(j.PublishedAt), nil))
	}
	return jobs, nil
}

func newJob(b Board, id, link, company, title, desc, location string, remote, internHint bool, posted, deadline *time.Time) IngestedJob {
	if location == "" {
		location = "Unspecified"
	}
	if posted == nil {
		now := time.Now().UTC()
		posted = &now
	}
	j := IngestedJob{
		ExternalID:      b.ExternalIDPrefix() + id,
		Source:          b.Source,
		SourceURL:       link,
		Type:            "JOB",
		CompanyName:     company,
		Title:           strings.TrimSpace(title),
		Description:     desc,
		Location:        location,
		IsRemote:        remote,
		JobType:         "FULL_TIME",
		ExperienceLevel: "ENTRY",
		PostedAt:        *posted,
		Deadline:        deadline,
	}
	if internHint || internRe.MatchString(title) {
		j.Type, j.JobType, j.ExperienceLevel = "INTERNSHIP", "INTERNSHIP", "INTERN"
	}
	return j
}

var (
	internRe  = regexp.MustCompile(`(?i)\b(intern|interns|internship|co-?op)\b`)
	studentRe = regexp.MustCompile(`(?i)\b(intern|interns|internship|co-?op|new grad(uate)?|graduate|entry[- ]level|junior|early career|university|campus|fresher|apprentice(ship)?)\b|\bSDE[- ]?(1|I)\b`)
	seniorRe  = regexp.MustCompile(`(?i)\b(recruit\w*|manager|senior|sr\.?|staff|principal|director|head|lead|mentor)\b`)
)

// IsStudentRole reports whether a posting title targets students or new graduates.
// ponytail: title keywords only; misses student roles with generic titles. Upgrade to
// description signals (graduation year, "currently enrolled") if coverage matters.
func IsStudentRole(title string) bool {
	return studentRe.MatchString(title) && !seniorRe.MatchString(title)
}

var (
	blockTagRe = regexp.MustCompile(`(?i)<\s*/?\s*(p|div|br|li|ul|ol|h[1-6]|tr)\b[^>]*>`)
	tagRe      = regexp.MustCompile(`<[^>]*>`)
	blankRe    = regexp.MustCompile(`\n\s*\n\s*`)
)

func htmlToText(s string) string {
	s = blockTagRe.ReplaceAllString(s, "\n")
	s = html.UnescapeString(tagRe.ReplaceAllString(s, ""))
	return strings.TrimSpace(blankRe.ReplaceAllString(s, "\n\n"))
}

func parseTime(s string) *time.Time {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

// SkillMatcher finds known skill names mentioned in free text.
type SkillMatcher struct {
	names []string
	res   []*regexp.Regexp
}

// ponytail: exact names only ("Go (Golang)" never matches "Go"); add an alias column to
// skills if extraction recall matters.
func NewSkillMatcher(known []string) *SkillMatcher {
	m := &SkillMatcher{}
	for _, name := range known {
		// Boundaries that treat "+", "#" and "." as part of a name (C++, C#, Node.js).
		re, err := regexp.Compile(`(?i)(?:^|[^\w+#.])` + regexp.QuoteMeta(name) + `(?:$|[^\w+#])`)
		if err != nil {
			continue
		}
		m.names = append(m.names, name)
		m.res = append(m.res, re)
	}
	return m
}

func (m *SkillMatcher) Match(text string) []string {
	found := []string{}
	for i, re := range m.res {
		if re.MatchString(text) {
			found = append(found, m.names[i])
		}
	}
	return found
}
