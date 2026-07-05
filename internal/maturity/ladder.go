// Package maturity implements the idea-maturity state machine: which stage
// transitions are legal, and which typed relations are allowed between stages.
// This is the core differentiator of the platform — no existing second-brain
// tool enforces a staged research pipeline, so this logic is entirely custom.
package maturity

import (
	"fmt"

	"github.com/kezlya/nobel/internal/domain"
)

// forwardLadder is the canonical happy-path progression of an idea.
var forwardLadder = []domain.Stage{
	domain.StageDraft,
	domain.StageHypothesis,
	domain.StageExperimentDesign,
	domain.StageExperimentLog,
	domain.StageWriteup,
}

// allowedTransitions maps a stage to the stages it may move to. The rules:
//   - advance one rung up the ladder,
//   - step back one rung to refine,
//   - a writeup forks into validated or invalidated,
//   - anything active can be archived.
var allowedTransitions = map[domain.Stage][]domain.Stage{
	domain.StageDraft:            {domain.StageHypothesis, domain.StageArchived},
	domain.StageHypothesis:       {domain.StageExperimentDesign, domain.StageDraft, domain.StageArchived},
	domain.StageExperimentDesign: {domain.StageExperimentLog, domain.StageHypothesis, domain.StageArchived},
	domain.StageExperimentLog:    {domain.StageWriteup, domain.StageExperimentDesign, domain.StageArchived},
	domain.StageWriteup:          {domain.StageValidated, domain.StageInvalidated, domain.StageExperimentLog, domain.StageArchived},
	domain.StageValidated:        {domain.StageArchived},
	domain.StageInvalidated:      {domain.StageArchived},
	domain.StageArchived:         {domain.StageDraft}, // un-archive back to the start of the ladder
}

// CanTransition reports whether a document may move from -> to.
func CanTransition(from, to domain.Stage) bool {
	for _, next := range allowedTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// NextStages returns the stages reachable from the given stage.
func NextStages(from domain.Stage) []domain.Stage {
	return append([]domain.Stage(nil), allowedTransitions[from]...)
}

// Validate returns a descriptive error if the transition is illegal.
func Validate(from, to domain.Stage) error {
	if !from.Valid() {
		return fmt.Errorf("unknown source stage %q", from)
	}
	if !to.Valid() {
		return fmt.Errorf("unknown target stage %q", to)
	}
	if from == to {
		return fmt.Errorf("document is already at stage %q", to)
	}
	if !CanTransition(from, to) {
		return fmt.Errorf("illegal transition %q -> %q; allowed: %v", from, to, allowedTransitions[from])
	}
	return nil
}

// relationGrammar constrains which typed relations may exist between documents
// at given stages. It encodes the Discourse-Graph-style grammar for the ladder.
// A relation not listed here is allowed only if it is one of the "free" cross-
// cutting relations (see freeRelations).
var relationGrammar = map[domain.RelationType]struct{ from, to domain.Stage }{
	domain.RelDerivedFrom: {from: domain.StageHypothesis, to: domain.StageDraft},
	domain.RelTests:       {from: domain.StageExperimentDesign, to: domain.StageHypothesis},
	domain.RelResultedIn:  {from: domain.StageExperimentLog, to: domain.StageWriteup},
}

// freeRelations may connect documents at any stages — they are the cross-cutting
// discourse edges rather than ladder-structural ones.
var freeRelations = map[domain.RelationType]bool{
	domain.RelSupports: true,
	domain.RelOpposes:  true,
	domain.RelInforms:  true,
	domain.RelRefines:  true,
}

// ValidateRelation checks that a typed edge is grammatical given the stages of
// its endpoints.
func ValidateRelation(rel domain.RelationType, sourceStage, targetStage domain.Stage) error {
	if !rel.Valid() {
		return fmt.Errorf("unknown relation type %q", rel)
	}
	if freeRelations[rel] {
		return nil
	}
	rule, ok := relationGrammar[rel]
	if !ok {
		return fmt.Errorf("relation %q has no defined grammar", rel)
	}
	if sourceStage != rule.from || targetStage != rule.to {
		return fmt.Errorf(
			"relation %q must go from a %q to a %q, got %q -> %q",
			rel, rule.from, rule.to, sourceStage, targetStage,
		)
	}
	return nil
}
