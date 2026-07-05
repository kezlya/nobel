package maturity

import (
	"testing"

	"github.com/kezlya/nobel/internal/domain"
)

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from, to domain.Stage
		want     bool
	}{
		{domain.StageDraft, domain.StageHypothesis, true},
		{domain.StageHypothesis, domain.StageExperimentDesign, true},
		{domain.StageHypothesis, domain.StageDraft, true},     // step back to refine
		{domain.StageWriteup, domain.StageValidated, true},    // fork
		{domain.StageWriteup, domain.StageInvalidated, true},  // fork
		{domain.StageDraft, domain.StageExperimentLog, false}, // cannot skip rungs
		{domain.StageValidated, domain.StageDraft, false},     // terminal, only archive
		{domain.StageArchived, domain.StageDraft, true},       // un-archive
	}
	for _, c := range cases {
		if got := CanTransition(c.from, c.to); got != c.want {
			t.Errorf("CanTransition(%q,%q)=%v want %v", c.from, c.to, got, c.want)
		}
	}
}

func TestValidateRejectsSameStage(t *testing.T) {
	if err := Validate(domain.StageDraft, domain.StageDraft); err == nil {
		t.Fatal("expected error for same-stage transition")
	}
}

func TestValidateRejectsUnknownStage(t *testing.T) {
	if err := Validate(domain.Stage("bogus"), domain.StageDraft); err == nil {
		t.Fatal("expected error for unknown source stage")
	}
}

func TestValidateRelationGrammar(t *testing.T) {
	// tests must go experiment_design -> hypothesis
	if err := ValidateRelation(domain.RelTests, domain.StageExperimentDesign, domain.StageHypothesis); err != nil {
		t.Errorf("expected valid tests relation, got %v", err)
	}
	if err := ValidateRelation(domain.RelTests, domain.StageDraft, domain.StageHypothesis); err == nil {
		t.Error("expected grammar violation for tests from a draft")
	}
	// free relations connect anything
	if err := ValidateRelation(domain.RelSupports, domain.StageWriteup, domain.StageDraft); err != nil {
		t.Errorf("expected free relation to be allowed, got %v", err)
	}
}
