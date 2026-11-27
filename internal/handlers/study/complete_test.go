package study

import (
	"testing"

	"github.com/csotherden/mandarin-flashcards/internal/models"
)

func TestSessionScore(t *testing.T) {
	if got := SessionScore(nil); got != 0 {
		t.Errorf("no grades = %v", got)
	}
	// Ending a 10-card session after one card at 72 scores 72, not 7.2.
	if got := SessionScore([]models.CardResult{{OverallScore: 72}}); got != 72 {
		t.Errorf("one graded card = %v, want 72", got)
	}
	if got := SessionScore([]models.CardResult{{OverallScore: 60}, {OverallScore: 90}}); got != 75 {
		t.Errorf("two cards = %v, want 75", got)
	}
}
