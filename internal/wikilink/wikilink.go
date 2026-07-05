// Package wikilink parses Obsidian-style [[wikilinks]] out of markdown bodies.
// Links are the edges of the knowledge graph; this package turns free text into
// the set of targets a document points at so the store can resolve them to
// document IDs and compute backlinks as the inverse set.
package wikilink

import (
	"regexp"
	"strings"
)

// linkPattern matches [[Target]] and [[Target|Alias]]. It intentionally does
// not match across newlines or nested brackets.
var linkPattern = regexp.MustCompile(`\[\[([^\[\]|]+)(?:\|([^\[\]]+))?\]\]`)

// Link is a single parsed reference. Target is the note name being linked to
// (used to resolve to a document); Alias is the optional display text.
type Link struct {
	Target string
	Alias  string
}

// Parse extracts every wikilink from a markdown body, in order of appearance.
// Duplicate targets are preserved so callers can count reference weight.
func Parse(body string) []Link {
	matches := linkPattern.FindAllStringSubmatch(body, -1)
	links := make([]Link, 0, len(matches))
	for _, m := range matches {
		target := strings.TrimSpace(m[1])
		if target == "" {
			continue
		}
		links = append(links, Link{Target: target, Alias: strings.TrimSpace(m[2])})
	}
	return links
}

// Targets returns the unique, order-preserved set of link targets in a body.
func Targets(body string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, l := range Parse(body) {
		if !seen[l.Target] {
			seen[l.Target] = true
			out = append(out, l.Target)
		}
	}
	return out
}

// Rename rewrites every [[old]] reference to [[new]] within a body, preserving
// aliases. This backs rename-safe link resolution: when a document's title
// changes, every body that references it is updated. Matching is exact on the
// target (case-sensitive), mirroring Obsidian's default behavior.
func Rename(body, old, updated string) string {
	return linkPattern.ReplaceAllStringFunc(body, func(match string) string {
		m := linkPattern.FindStringSubmatch(match)
		if strings.TrimSpace(m[1]) != old {
			return match
		}
		if alias := strings.TrimSpace(m[2]); alias != "" {
			return "[[" + updated + "|" + alias + "]]"
		}
		return "[[" + updated + "]]"
	})
}
