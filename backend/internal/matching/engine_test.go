package matching

import (
	"testing"
	"time"
)

func TestMatchingEngine_HighMatch(t *testing.T) {
	engine := NewEngine(DefaultWeights)

	futureDeadline := time.Now().Add(14 * 24 * time.Hour)

	candidate := &CandidateProfile{
		ID:                 "cand-1",
		Degree:             "B.Tech",
		Branch:             "Computer Science",
		GraduationYear:     2026,
		CGPA:               8.5,
		Skills:             []string{"Python", "PyTorch", "Machine Learning", "FastAPI"},
		TargetRoles:        []string{"Machine Learning Engineer", "AI Engineer"},
		PreferredLocations: []string{"Bangalore", "Remote"},
		WorkPreference:     "ANY",
		ProjectSkills:      []string{"PyTorch", "Docker"},
	}

	opp := &OpportunityItem{
		ID:                "opp-1",
		Title:             "Machine Learning Engineer Intern",
		CompanyName:       "CognitiveScale",
		Location:          "Bangalore, India",
		IsRemote:          false,
		RequiredSkills:    []string{"Python", "PyTorch", "Machine Learning"},
		OptionalSkills:    []string{"Docker"},
		EligibleDegrees:   []string{"B.Tech", "M.Tech"},
		EligibleGradYears: []int{2025, 2026},
		MinCGPA:           7.5,
		Deadline:          &futureDeadline,
		Status:            "ACTIVE",
	}

	result, eligible := engine.Evaluate(candidate, opp, 0.90)
	if !eligible {
		t.Fatalf("Expected candidate to be eligible for opportunity")
	}

	if result.MatchPercentage < 80 {
		t.Errorf("Expected high match score (>=80%%), got %d%%", result.MatchPercentage)
	}

	if len(result.MatchedReasons) == 0 {
		t.Errorf("Expected matched reasons to be populated")
	}

	t.Logf("High match test passed: %d%% match. Reasons: %v", result.MatchPercentage, result.MatchedReasons)
}

func TestMatchingEngine_LowMatch(t *testing.T) {
	engine := NewEngine(DefaultWeights)

	futureDeadline := time.Now().Add(14 * 24 * time.Hour)

	candidate := &CandidateProfile{
		ID:                 "cand-2",
		Degree:             "B.Tech",
		Branch:             "Civil Engineering",
		GraduationYear:     2026,
		CGPA:               6.8,
		Skills:             []string{"AutoCAD", "Surveying"},
		TargetRoles:        []string{"Civil Engineer"},
		PreferredLocations: []string{"Delhi"},
		WorkPreference:     "ONSITE",
	}

	opp := &OpportunityItem{
		ID:                "opp-2",
		Title:             "Senior Golang Distributed Systems Intern",
		CompanyName:       "CloudScale",
		Location:          "Bangalore, India",
		IsRemote:          false,
		RequiredSkills:    []string{"Go", "Kubernetes", "PostgreSQL", "Kafka"},
		EligibleDegrees:   []string{"B.Tech"},
		EligibleGradYears: []int{2026},
		MinCGPA:           7.5,
		Deadline:          &futureDeadline,
		Status:            "ACTIVE",
	}

	result, eligible := engine.Evaluate(candidate, opp, 0.10)
	if !eligible {
		t.Fatalf("Expected candidate to pass basic eligibility")
	}

	if result.MatchPercentage > 40 {
		t.Errorf("Expected low match score (<=40%%), got %d%%", result.MatchPercentage)
	}

	if len(result.MissingRequirements) == 0 {
		t.Errorf("Expected missing requirements to highlight skill and location gaps")
	}

	t.Logf("Low match test passed: %d%% match. Missing: %v", result.MatchPercentage, result.MissingRequirements)
}

func TestMatchingEngine_DeterministicPruning(t *testing.T) {
	engine := NewEngine(DefaultWeights)

	candidate := &CandidateProfile{
		ID:             "cand-3",
		Degree:         "B.Tech",
		GraduationYear: 2026,
	}

	// 1. Expired deadline
	pastDeadline := time.Now().Add(-2 * time.Hour)
	expiredOpp := &OpportunityItem{
		ID:       "opp-exp",
		Status:   "ACTIVE",
		Deadline: &pastDeadline,
	}
	if _, eligible := engine.Evaluate(candidate, expiredOpp, 0.8); eligible {
		t.Errorf("Expected expired opportunity to be pruned")
	}

	// 2. Incompatible Graduation Year
	futureDeadline := time.Now().Add(24 * time.Hour)
	incompatibleYearOpp := &OpportunityItem{
		ID:                "opp-yr",
		Status:            "ACTIVE",
		Deadline:          &futureDeadline,
		EligibleGradYears: []int{2024}, // candidate is 2026
	}
	if _, eligible := engine.Evaluate(candidate, incompatibleYearOpp, 0.8); eligible {
		t.Errorf("Expected incompatible graduation year opportunity to be pruned")
	}
}
