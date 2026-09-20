package matching

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type CandidateProfile struct {
	ID                 string
	Degree             string
	Branch             string
	GraduationYear     int
	CGPA               float64
	Skills             []string
	TargetRoles        []string
	PreferredLocations []string
	WorkPreference     string
	ProjectSkills      []string
	AppliedJobIDs      map[string]bool
}

type OpportunityItem struct {
	ID                string
	Title             string
	CompanyName       string
	Description       string
	Location          string
	IsRemote          bool
	RequiredSkills    []string
	OptionalSkills    []string
	EligibleDegrees   []string
	EligibleGradYears []int
	MinCGPA           float64
	Deadline          *time.Time
	Status            string
}

type MatchResult struct {
	OpportunityID       string   `json:"opportunity_id"`
	MatchPercentage     int      `json:"match_percentage"`
	SkillScore          float64  `json:"skill_score"`
	RoleScore           float64  `json:"role_score"`
	EligibilityScore    float64  `json:"eligibility_score"`
	LocationScore       float64  `json:"location_score"`
	ProjectScore        float64  `json:"project_score"`
	SemanticScore       float64  `json:"semantic_score"`
	MatchedReasons      []string `json:"matched_reasons"`
	MissingRequirements []string `json:"missing_requirements"`
}

type Engine struct {
	weights Weights
}

func NewEngine(weights Weights) *Engine {
	return &Engine{weights: weights}
}

// Evaluate performs Stage 1 deterministic filtering and Stage 2 multi-signal scoring
func (e *Engine) Evaluate(candidate *CandidateProfile, opp *OpportunityItem, semanticSimilarity float64) (*MatchResult, bool) {
	// ==========================================
	// STAGE 1: Deterministic Hard Constraints
	// ==========================================
	if opp.Status != "ACTIVE" {
		return nil, false
	}

	if opp.Deadline != nil && time.Now().After(*opp.Deadline) {
		return nil, false
	}

	if candidate.AppliedJobIDs != nil && candidate.AppliedJobIDs[opp.ID] {
		return nil, false
	}

	// Degree filter
	if len(opp.EligibleDegrees) > 0 {
		degreeMatched := false
		for _, d := range opp.EligibleDegrees {
			if strings.EqualFold(d, "Any Degree") || strings.EqualFold(d, candidate.Degree) || strings.Contains(strings.ToLower(candidate.Degree), strings.ToLower(d)) {
				degreeMatched = true
				break
			}
		}
		if !degreeMatched {
			return nil, false
		}
	}

	// Graduation Year filter
	if len(opp.EligibleGradYears) > 0 && candidate.GraduationYear > 0 {
		gradYearMatched := false
		for _, yr := range opp.EligibleGradYears {
			if yr == candidate.GraduationYear {
				gradYearMatched = true
				break
			}
		}
		if !gradYearMatched {
			return nil, false
		}
	}

	// ==========================================
	// STAGE 2: Multi-Signal Scoring & Explainability
	// ==========================================
	var matchedReasons []string
	var missingReqs []string

	// 1. Skill Score
	skillScore := 0.0
	candidateSkillMap := make(map[string]bool)
	for _, s := range candidate.Skills {
		candidateSkillMap[strings.ToLower(strings.TrimSpace(s))] = true
	}
	for _, ps := range candidate.ProjectSkills {
		candidateSkillMap[strings.ToLower(strings.TrimSpace(ps))] = true
	}

	reqMatched := 0
	if len(opp.RequiredSkills) > 0 {
		for _, rs := range opp.RequiredSkills {
			cleanRS := strings.ToLower(strings.TrimSpace(rs))
			if candidateSkillMap[cleanRS] {
				reqMatched++
				matchedReasons = append(matchedReasons, fmt.Sprintf("Verified skill: %s", rs))
			} else {
				missingReqs = append(missingReqs, fmt.Sprintf("%s required but not in profile", rs))
			}
		}
		skillScore = float64(reqMatched) / float64(len(opp.RequiredSkills))
	} else {
		skillScore = 1.0
	}

	// 2. Role Score (Target Roles vs Opportunity Title)
	roleScore := 0.0
	if len(candidate.TargetRoles) > 0 {
		for _, tr := range candidate.TargetRoles {
			score := calculateStringSimilarity(tr, opp.Title)
			if score > roleScore {
				roleScore = score
			}
		}
		if roleScore > 0.4 {
			matchedReasons = append(matchedReasons, fmt.Sprintf("Role alignment with title '%s'", opp.Title))
		}
	} else {
		roleScore = 0.5
	}

	// 3. Eligibility Score (CGPA & Degree)
	eligScore := 1.0
	if opp.MinCGPA > 0 {
		if candidate.CGPA >= opp.MinCGPA {
			matchedReasons = append(matchedReasons, fmt.Sprintf("Meets CGPA requirement (%.1f >= %.1f)", candidate.CGPA, opp.MinCGPA))
		} else {
			eligScore = math.Max(0, candidate.CGPA/opp.MinCGPA)
			missingReqs = append(missingReqs, fmt.Sprintf("Minimum CGPA %.1f required (Current: %.1f)", opp.MinCGPA, candidate.CGPA))
		}
	}

	// 4. Location Score
	locScore := 0.0
	if opp.IsRemote || strings.EqualFold(candidate.WorkPreference, "REMOTE") {
		locScore = 1.0
		matchedReasons = append(matchedReasons, "Remote work opportunity")
	} else if len(candidate.PreferredLocations) > 0 {
		for _, pl := range candidate.PreferredLocations {
			if strings.Contains(strings.ToLower(opp.Location), strings.ToLower(strings.TrimSpace(pl))) {
				locScore = 1.0
				matchedReasons = append(matchedReasons, fmt.Sprintf("Location '%s' matches preference", pl))
				break
			}
		}
		if locScore == 0 {
			locScore = 0.3
			missingReqs = append(missingReqs, fmt.Sprintf("Located in %s (Outside preferred locations)", opp.Location))
		}
	} else {
		locScore = 0.8
	}

	// 5. Project Score
	projectScore := 0.0
	if len(candidate.ProjectSkills) > 0 && len(opp.RequiredSkills) > 0 {
		projMatched := 0
		for _, rs := range opp.RequiredSkills {
			for _, ps := range candidate.ProjectSkills {
				if strings.EqualFold(rs, ps) {
					projMatched++
					matchedReasons = append(matchedReasons, fmt.Sprintf("Project validated experience: %s", rs))
					break
				}
			}
		}
		projectScore = math.Min(1.0, float64(projMatched)/float64(len(opp.RequiredSkills)))
	}

	// 6. Semantic Similarity
	semScore := semanticSimilarity
	if semScore <= 0 {
		// If vector embedding not provided or offline, redistribute weight gracefully
		semScore = (skillScore + roleScore + locScore) / 3.0
	}

	// Final weighted score calculation
	totalScore := (e.weights.Skill * skillScore) +
		(e.weights.Role * roleScore) +
		(e.weights.Eligibility * eligScore) +
		(e.weights.Location * locScore) +
		(e.weights.Project * projectScore) +
		(e.weights.Semantic * semScore)

	percentage := int(math.Round(totalScore * 100))
	if percentage > 100 {
		percentage = 100
	} else if percentage < 0 {
		percentage = 0
	}

	return &MatchResult{
		OpportunityID:       opp.ID,
		MatchPercentage:     percentage,
		SkillScore:          skillScore,
		RoleScore:           roleScore,
		EligibilityScore:    eligScore,
		LocationScore:       locScore,
		ProjectScore:        projectScore,
		SemanticScore:       semScore,
		MatchedReasons:      matchedReasons,
		MissingRequirements: missingReqs,
	}, true
}

func calculateStringSimilarity(target, title string) float64 {
	tLower := strings.ToLower(target)
	titleLower := strings.ToLower(title)

	if strings.Contains(titleLower, tLower) || strings.Contains(tLower, titleLower) {
		return 1.0
	}

	targetWords := strings.Fields(tLower)
	titleWords := strings.Fields(titleLower)

	matches := 0
	for _, tw := range targetWords {
		for _, titw := range titleWords {
			if tw == titw {
				matches++
				break
			}
		}
	}

	if len(targetWords) == 0 {
		return 0.0
	}
	return float64(matches) / float64(len(targetWords))
}
