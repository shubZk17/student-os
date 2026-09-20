package matching

type Weights struct {
	Skill       float64
	Role        float64
	Eligibility float64
	Location    float64
	Project     float64
	Semantic    float64
}

// DefaultWeights returns the production weights as specified in Section 12 of the PRD
var DefaultWeights = Weights{
	Skill:       0.40,
	Role:        0.20,
	Eligibility: 0.15,
	Location:    0.10,
	Project:     0.05,
	Semantic:    0.10,
}
