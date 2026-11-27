package streak

import (
	"testing"
	"time"
)

func TestRuns(t *testing.T) {
	cases := []struct {
		name             string
		days             []string
		current, longest int
		last             string
	}{
		{"single", []string{"2026-11-27"}, 1, 1, "2026-11-27"},
		{"repeats and order", []string{"2026-11-27", "2026-11-25", "2026-11-26", "2026-11-27"}, 3, 3, "2026-11-27"},
		{"gap keeps longest", []string{"2026-11-01", "2026-11-02", "2026-11-03", "2026-11-10", "2026-11-11"}, 2, 3, "2026-11-11"},
		{"across months and years", []string{"2025-12-31", "2026-01-01", "2026-01-02"}, 3, 3, "2026-01-02"},
		{"longer than 120 days", spanDays("2026-01-01", 200), 200, 200, "2026-07-19"},
	}
	for _, c := range cases {
		current, longest, last := runs(c.days)
		if current != c.current || longest != c.longest || last != c.last {
			t.Errorf("%s: runs = (%d, %d, %s), want (%d, %d, %s)", c.name, current, longest, last, c.current, c.longest, c.last)
		}
	}
}

func TestValidTimezone(t *testing.T) {
	for _, name := range []string{"America/New_York", "Asia/Shanghai", "UTC"} {
		if !ValidTimezone(name) {
			t.Errorf("%s should be valid", name)
		}
	}
	for _, name := range []string{"", "Mars/Olympus", "../etc/passwd"} {
		if ValidTimezone(name) {
			t.Errorf("%q should be invalid", name)
		}
	}
}

func spanDays(start string, n int) []string {
	day, _ := time.Parse(dayLayout, start)
	out := make([]string, n)
	for i := range out {
		out[i] = day.AddDate(0, 0, i).Format(dayLayout)
	}
	return out
}
