// Package xp awards experience points for graded cards.
//
// Each grade lands in a tier. A word earns, per day (the user's own calendar day), the XP of the
// best tier it reached: a retry that reaches a higher tier tops up the difference, one that
// doesn't earns nothing. Words, not cards, are the unit, so duplicate cards don't double up.
// The photo-to-hanzi match drill earns half, since recognizing is easier than saying.
package xp

import (
	"context"
	"fmt"
	"time"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/ent/xpaward"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/streak"
	"github.com/google/uuid"
)

// Tier names, lowest to highest. They are also the labels the card back shows.
const (
	None    = ""
	Keep    = "keep"    // graded, under CloseScore
	Close   = "close"   // CloseScore and up
	Nice    = "nice"    // NiceScore and up
	Perfect = "perfect" // every syllable right and PerfectScore and up
)

// Score cutoffs. NiceScore matches the "slipped" threshold used on Today.
const (
	CloseScore   = 60
	NiceScore    = 85
	PerfectScore = 95
)

var speakXP = map[string]int{None: 0, Keep: 2, Close: 5, Nice: 8, Perfect: 12}

// Points is the XP a tier is worth in a mode. Match earns half, rounded up.
func Points(tier, mode string) int {
	p := speakXP[tier]
	if mode == models.ModePictureHanziMatch {
		return (p + 1) / 2
	}
	return p
}

// SpeakTier grades a spoken take. Takes that couldn't be graded earn nothing, so silence or
// noise can't farm XP.
func SpeakTier(assessable bool, score float64, syllables []models.SyllableView) string {
	if !assessable {
		return None
	}
	if score >= PerfectScore && allOK(syllables) {
		return Perfect
	}
	switch {
	case score >= NiceScore:
		return Nice
	case score >= CloseScore:
		return Close
	default:
		return Keep
	}
}

func allOK(syllables []models.SyllableView) bool {
	if len(syllables) == 0 {
		return false
	}
	for _, s := range syllables {
		if s.Status != "ok" {
			return false
		}
	}
	return true
}

// MatchTier grades a match: a wrong pick earns nothing, a first-try match is perfect, and each
// earlier miss lowers the score (and so the tier) of the eventual match.
func MatchTier(correct bool, mismatches int, score float64) string {
	switch {
	case !correct:
		return None
	case mismatches == 0:
		return Perfect
	case score >= NiceScore:
		return Nice
	case score >= CloseScore:
		return Close
	default:
		return Keep
	}
}

// Rank orders tiers so a retry can be compared with the day's best.
func Rank(tier string) int {
	switch tier {
	case Keep:
		return 1
	case Close:
		return 2
	case Nice:
		return 3
	case Perfect:
		return 4
	}
	return 0
}

// Result is what one grade earned.
type Result struct {
	Earned int    // XP added to the total by this grade
	Tier   string // the tier this grade reached
	Total  int    // the user's XP total afterwards
}

// Award records a graded log's tier for its word and day, adds any top-up to the user's total,
// and stamps what was earned onto the log, in one transaction. The user's row is locked for the
// duration, so concurrent grades for one user are applied one at a time.
func Award(ctx context.Context, db *ent.Client, userID uuid.UUID, log *ent.CardAttemptLog, word, tier string, score float64) (Result, error) {
	tx, err := db.Tx(ctx)
	if err != nil {
		return Result{}, err
	}
	res, err := award(ctx, tx, userID, log, word, tier, score)
	if err != nil {
		_ = tx.Rollback()
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return res, nil
}

func award(ctx context.Context, tx *ent.Tx, userID uuid.UUID, log *ent.CardAttemptLog, word, tier string, score float64) (Result, error) {
	u, err := tx.User.Query().Where(user.ID(userID)).ForUpdate().Only(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("lock user: %w", err)
	}
	res := Result{Tier: tier, Total: u.XpTotal}
	points := Points(tier, string(log.Mode))
	day := log.CreatedAt.In(streak.Location(u)).Format("2006-01-02")

	if word != "" && points > 0 {
		prior, err := tx.XPAward.Query().
			Where(xpaward.UserID(userID), xpaward.Word(word), xpaward.Day(day)).
			Only(ctx)
		switch {
		case ent.IsNotFound(err):
			_, err = tx.XPAward.Create().
				SetUserID(userID).SetWord(word).SetDay(day).
				SetXp(points).SetTier(tier).SetBestScore(score).
				Save(ctx)
			if err != nil {
				return Result{}, fmt.Errorf("create award: %w", err)
			}
			res.Earned = points
		case err != nil:
			return Result{}, fmt.Errorf("read award: %w", err)
		case points > prior.Xp:
			_, err = prior.Update().SetXp(points).SetTier(tier).SetBestScore(max(score, prior.BestScore)).Save(ctx)
			if err != nil {
				return Result{}, fmt.Errorf("raise award: %w", err)
			}
			res.Earned = points - prior.Xp
		}
	}
	if res.Earned > 0 {
		if err := tx.User.UpdateOneID(userID).AddXpTotal(res.Earned).Exec(ctx); err != nil {
			return Result{}, fmt.Errorf("add xp: %w", err)
		}
		res.Total += res.Earned
	}
	if err := tx.CardAttemptLog.UpdateOneID(log.ID).SetXpEarned(res.Earned).SetXpTier(tier).Exec(ctx); err != nil {
		return Result{}, fmt.Errorf("stamp log: %w", err)
	}
	return res, nil
}

// Today is the XP a user has earned on their current calendar day.
func Today(ctx context.Context, db *ent.Client, u *ent.User, now time.Time) (int, error) {
	day := now.In(streak.Location(u)).Format("2006-01-02")
	points, err := db.XPAward.Query().
		Where(xpaward.UserID(u.ID), xpaward.Day(day)).
		Select(xpaward.FieldXp).
		Ints(ctx)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, p := range points {
		total += p
	}
	return total, nil
}
