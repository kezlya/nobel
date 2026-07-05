// Package domain holds the core types of the Idea Maturity Platform: documents
// (nodes), the maturity stages they move through, and the typed relations
// (edges) that connect them. It has no dependencies on storage or transport so
// it can be reused across the HTTP layer, background jobs, and tests.
package domain

import "time"

// Stage is a point on the idea-maturity ladder. Every document sits at exactly
// one stage. The ordered progression is the platform's differentiator: no
// off-the-shelf second-brain tool enforces a staged pipeline.
type Stage string

const (
	StageDraft            Stage = "draft"             // a raw idea, not yet a testable claim
	StageHypothesis       Stage = "hypothesis"        // a falsifiable claim derived from a draft
	StageExperimentDesign Stage = "experiment_design" // a plan that tests a hypothesis
	StageExperimentLog    Stage = "experiment_log"    // the record of running an experiment
	StageWriteup          Stage = "writeup"           // conclusions synthesized from logs
	StageValidated        Stage = "validated"         // terminal: hypothesis held up
	StageInvalidated      Stage = "invalidated"       // terminal: hypothesis did not hold
	StageArchived         Stage = "archived"          // terminal: parked, out of the active graph
)

// AllStages lists every known stage in ladder order.
func AllStages() []Stage {
	return []Stage{
		StageDraft, StageHypothesis, StageExperimentDesign, StageExperimentLog,
		StageWriteup, StageValidated, StageInvalidated, StageArchived,
	}
}

// Valid reports whether s is a known stage.
func (s Stage) Valid() bool {
	for _, k := range AllStages() {
		if s == k {
			return true
		}
	}
	return false
}

// RelationType is the semantic label on an edge between two documents. The set
// is adapted from the Discourse Graphs ontology (supports/opposes/informs/
// derivedFrom) plus ladder-specific relations (tests/refines/resultedIn).
type RelationType string

const (
	RelSupports    RelationType = "supports"
	RelOpposes     RelationType = "opposes"
	RelInforms     RelationType = "informs"
	RelDerivedFrom RelationType = "derivedFrom"
	RelTests       RelationType = "tests"
	RelRefines     RelationType = "refines"
	RelResultedIn  RelationType = "resultedIn"
)

// AllRelationTypes lists every known relation type.
func AllRelationTypes() []RelationType {
	return []RelationType{
		RelSupports, RelOpposes, RelInforms, RelDerivedFrom,
		RelTests, RelRefines, RelResultedIn,
	}
}

// Valid reports whether r is a known relation type.
func (r RelationType) Valid() bool {
	for _, k := range AllRelationTypes() {
		if r == k {
			return true
		}
	}
	return false
}

// Document is a node in the knowledge graph: a markdown file plus metadata.
type Document struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	BodyMarkdown string    `json:"body_markdown"`
	Stage        Stage     `json:"stage"`
	AuthorID     string    `json:"author_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Link is a directed, typed edge from one document to another. Backlinks are
// simply the set of links whose Target is a given document — they are computed,
// never stored separately.
type Link struct {
	ID           string       `json:"id"`
	SourceID     string       `json:"source_id"`
	TargetID     string       `json:"target_id"`
	RelationType RelationType `json:"relation_type"`
	CreatedAt    time.Time    `json:"created_at"`
}

// StageTransition is an audit record of a document moving between stages.
type StageTransition struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	FromStage  Stage     `json:"from_stage"`
	ToStage    Stage     `json:"to_stage"`
	ActorID    string    `json:"actor_id"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}
