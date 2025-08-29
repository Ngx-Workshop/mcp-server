package criteria

// Domain: Criterion, Score, Evidence
// This file defines the core domain models for criteria evaluation including
// criterion definitions, scoring mechanisms, and evidence collection structures

type Criterion struct {
	Key    string
	Value  float64
	Weight float64
	Source string // quiz, rubric, behavior, etc.
}

type Criteria struct {
	LearnerID string
	CourseID  string
	Items     []Criterion
}
