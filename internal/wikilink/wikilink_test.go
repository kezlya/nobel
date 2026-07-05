package wikilink

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	body := "See [[Attention Is All You Need]] and [[GPT-2|the paper]]. Also [[Attention Is All You Need]] again."
	got := Parse(body)
	want := []Link{
		{Target: "Attention Is All You Need"},
		{Target: "GPT-2", Alias: "the paper"},
		{Target: "Attention Is All You Need"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse()=%+v want %+v", got, want)
	}
}

func TestTargetsDeduplicates(t *testing.T) {
	body := "[[A]] [[B]] [[A]] [[C|c]]"
	got := Targets(body)
	want := []string{"A", "B", "C"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Targets()=%v want %v", got, want)
	}
}

func TestRename(t *testing.T) {
	body := "link to [[Old Name]] and [[Old Name|alias]] and [[Other]]"
	got := Rename(body, "Old Name", "New Name")
	want := "link to [[New Name]] and [[New Name|alias]] and [[Other]]"
	if got != want {
		t.Errorf("Rename()=%q want %q", got, want)
	}
}

func TestParseIgnentsEmpty(t *testing.T) {
	if got := Parse("[[]] and [[   ]]"); len(got) != 0 {
		t.Errorf("expected no links, got %+v", got)
	}
}
