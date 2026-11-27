package home

import (
	"testing"
	"time"
)

func TestStreak(t *testing.T) {
	now := time.Date(2026, 11, 27, 9, 0, 0, 0, time.UTC)
	days := map[string]bool{"2026-11-27": true, "2026-11-26": true, "2026-11-25": true, "2026-11-23": true}
	if got := streak(days, now); got != 3 {
		t.Errorf("streak with today = %d, want 3", got)
	}
	delete(days, "2026-11-27")
	if got := streak(days, now); got != 2 {
		t.Errorf("streak ending yesterday = %d, want 2 (today not practiced yet)", got)
	}
	if got := streak(map[string]bool{"2026-11-20": true}, now); got != 0 {
		t.Errorf("broken streak = %d, want 0", got)
	}
}
