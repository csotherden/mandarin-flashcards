package study

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPrioritize(t *testing.T) {
	weakID := uuid.New()
	unseenID := uuid.New()
	strongID := uuid.New()
	older := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-time.Minute)
	weak := 40.0
	strong := 90.0
	alsoStrong := 90.0

	olderStrong := uuid.New()
	cards := []PracticeStat{
		{ID: strongID, Average: &strong, LastPracticed: &newer},
		{ID: unseenID},
		{ID: weakID, Average: &weak, LastPracticed: &newer},
		{ID: olderStrong, Average: &alsoStrong, LastPracticed: &older},
	}

	got := Prioritize(cards, 0)
	if len(got) != 4 {
		t.Fatalf("len %d", len(got))
	}
	if got[0] != weakID {
		t.Fatalf("weak card first, got %s", got[0])
	}
	if got[1] != unseenID {
		t.Fatalf("unseen card second, got %s", got[1])
	}
	if got[2] != olderStrong {
		t.Fatalf("older strong card before recent strong card, got %s", got[2])
	}

	top := Prioritize(cards, 1)
	if len(top) != 1 || top[0] != weakID {
		t.Fatalf("limit 1 = %v", top)
	}
}
