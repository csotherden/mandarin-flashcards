package study

import (
	"time"

	"github.com/google/uuid"
)

// PracticeStat is one card's history used to choose the next drill.
type PracticeStat struct {
	ID            uuid.UUID
	Average       *float64
	LastPracticed *time.Time
}

// Prioritize orders cards so weaker and less recently practiced ones come first.
// A card with no history is treated as a neutral 50, so real misses (under 50)
// surface first and unseen cards still come ahead of words the learner already
// scores well on. limit <= 0 returns every card.
func Prioritize(cards []PracticeStat, limit int) []uuid.UUID {
	sorted := append([]PracticeStat(nil), cards...)
	for i := 1; i < len(sorted); i++ {
		item := sorted[i]
		j := i
		for j > 0 && lessStat(item, sorted[j-1]) {
			sorted[j] = sorted[j-1]
			j--
		}
		sorted[j] = item
	}
	if limit <= 0 || limit > len(sorted) {
		limit = len(sorted)
	}
	out := make([]uuid.UUID, limit)
	for i := 0; i < limit; i++ {
		out[i] = sorted[i].ID
	}
	return out
}

func lessStat(a, b PracticeStat) bool {
	av, bv := average(a), average(b)
	if av != bv {
		return av < bv
	}
	return earlier(a.LastPracticed, b.LastPracticed)
}

func average(card PracticeStat) float64 {
	if card.Average == nil {
		return 50
	}
	return *card.Average
}

func earlier(a, b *time.Time) bool {
	switch {
	case a == nil && b == nil:
		return false
	case a == nil:
		return true
	case b == nil:
		return false
	default:
		return a.Before(*b)
	}
}
