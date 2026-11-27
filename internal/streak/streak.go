// Package streak keeps each user's practice streak as stored counters instead of recounting
// history on every read.
//
// A practice day is a calendar day, in the user's own saved timezone, with at least one graded
// card. Record moves the counters forward when a card is graded. Check settles a lapse lazily:
// nothing runs for a user until someone reads their streak, so there is no sweep over all users.
// Both take the user id explicitly, so the same calls serve the signed-in user and, later,
// someone viewing another user's profile.
package streak

import (
	"context"
	"fmt"
	"sort"
	"time"
	_ "time/tzdata" // the runtime image has no zoneinfo; embed it so LoadLocation always works

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardattemptlog"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/google/uuid"
)

const dayLayout = "2006-01-02"

// Status is a user's streak as of now.
type Status struct {
	// Current is the live streak: 0 once a whole day has passed without practice.
	Current int
	Longest int
	// Today is true when today already counts; Current > 0 && !Today means the streak is
	// alive but needs practice today to continue.
	Today bool
}

// ValidTimezone reports whether name is an IANA zone this server can resolve.
func ValidTimezone(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	_, err := time.LoadLocation(name)
	return err == nil
}

// Location is the user's saved zone, or UTC when none is saved yet.
func Location(u *ent.User) *time.Location {
	if u.Timezone != "" {
		if loc, err := time.LoadLocation(u.Timezone); err == nil {
			return loc
		}
	}
	return time.UTC
}

// Check returns the user's streak, settling a lapse (and, the first time, backfilling the
// counters from their full practice history).
func Check(ctx context.Context, db *ent.Client, userID uuid.UUID, now time.Time) (Status, error) {
	u, err := db.User.Get(ctx, userID)
	if err != nil {
		return Status{}, err
	}
	if u.StreakLastDay == nil {
		if u, err = backfill(ctx, db, u); err != nil {
			return Status{}, err
		}
	}
	loc := Location(u)
	today := now.In(loc).Format(dayLayout)
	yesterday := now.In(loc).AddDate(0, 0, -1).Format(dayLayout)
	status := Status{Current: u.StreakCurrent, Longest: u.StreakLongest}
	if u.StreakLastDay == nil {
		return Status{Longest: u.StreakLongest}, nil
	}
	last := *u.StreakLastDay
	switch {
	case last == today:
		status.Today = true
	case last == yesterday:
	default:
		status.Current = 0
		if u.StreakCurrent != 0 {
			// Conditional on the day read above, so a grade recorded meanwhile is not undone.
			err := db.User.Update().
				Where(user.ID(userID), user.StreakLastDay(last)).
				SetStreakCurrent(0).
				Exec(ctx)
			if err != nil {
				return Status{}, fmt.Errorf("settle lapsed streak: %w", err)
			}
		}
	}
	return status, nil
}

// Record counts a graded card at the given time toward the user's streak. Safe to call for
// every grade: only the first one of a day changes anything.
func Record(ctx context.Context, db *ent.Client, userID uuid.UUID, at time.Time) error {
	u, err := db.User.Get(ctx, userID)
	if err != nil {
		return err
	}
	if u.StreakLastDay == nil {
		// First practice since streaks were stored: the card history (including this grade,
		// if it was a card) gives the right counters.
		if u, err = backfill(ctx, db, u); err != nil {
			return err
		}
		if u.StreakLastDay == nil {
			// No card history (say, a tutorial take came first): this is day one.
			loc := Location(u)
			return db.User.Update().
				Where(user.ID(userID), user.StreakLastDayIsNil()).
				SetStreakCurrent(1).
				SetStreakLongest(max(1, u.StreakLongest)).
				SetStreakLastDay(at.In(loc).Format(dayLayout)).
				Exec(ctx)
		}
		return nil
	}
	loc := Location(u)
	today := at.In(loc).Format(dayLayout)
	if *u.StreakLastDay == today {
		return nil
	}
	yesterday := at.In(loc).AddDate(0, 0, -1).Format(dayLayout)
	// Each update is conditional on the stored day, so concurrent grades count today once.
	n, err := db.User.Update().
		Where(user.ID(userID), user.StreakLastDay(yesterday)).
		AddStreakCurrent(1).
		SetStreakLastDay(today).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("extend streak: %w", err)
	}
	if n == 0 {
		_, err = db.User.Update().
			Where(user.ID(userID), user.StreakLastDayNEQ(today)).
			SetStreakCurrent(1).
			SetStreakLastDay(today).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("restart streak: %w", err)
		}
	}
	return raiseLongest(ctx, db, userID)
}

func raiseLongest(ctx context.Context, db *ent.Client, userID uuid.UUID) error {
	u, err := db.User.Get(ctx, userID)
	if err != nil {
		return err
	}
	if u.StreakCurrent <= u.StreakLongest {
		return nil
	}
	return db.User.Update().
		Where(user.ID(userID), user.StreakLongestLT(u.StreakCurrent)).
		SetStreakLongest(u.StreakCurrent).
		Exec(ctx)
}

// backfill computes the counters from every graded card the user has, once. A user with no
// practice keeps a nil last day (the next read repeats one empty query, nothing more).
func backfill(ctx context.Context, db *ent.Client, u *ent.User) (*ent.User, error) {
	var rows []struct {
		CreatedAt time.Time `json:"created_at"`
	}
	err := db.CardAttemptLog.Query().
		Where(cardattemptlog.UserID(u.ID)).
		Select(cardattemptlog.FieldCreatedAt).
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("read practice history: %w", err)
	}
	if len(rows) == 0 {
		return u, nil
	}
	loc := Location(u)
	days := make([]string, 0, len(rows))
	for _, row := range rows {
		days = append(days, row.CreatedAt.In(loc).Format(dayLayout))
	}
	current, longest, last := runs(days)
	// Conditional on still being unset: if a concurrent call backfilled first, keep theirs.
	err = db.User.Update().
		Where(user.ID(u.ID), user.StreakLastDayIsNil()).
		SetStreakCurrent(current).
		SetStreakLongest(max(longest, u.StreakLongest)).
		SetStreakLastDay(last).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("store streak: %w", err)
	}
	return db.User.Get(ctx, u.ID)
}

// runs takes practice days (YYYY-MM-DD, any order, repeats allowed) and returns the run of
// consecutive days ending on the latest one, the longest run, and the latest day.
func runs(days []string) (current, longest int, last string) {
	set := map[string]bool{}
	for _, d := range days {
		set[d] = true
	}
	sorted := make([]string, 0, len(set))
	for d := range set {
		sorted = append(sorted, d)
	}
	sort.Strings(sorted)
	run := 0
	var prev time.Time
	for i, d := range sorted {
		day, err := time.Parse(dayLayout, d)
		if err != nil {
			continue
		}
		if i > 0 && day.Equal(prev.AddDate(0, 0, 1)) {
			run++
		} else {
			run = 1
		}
		longest = max(longest, run)
		prev = day
	}
	return run, longest, sorted[len(sorted)-1]
}
