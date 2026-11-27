package xp

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardattemptlog"
	"github.com/csotherden/mandarin-flashcards/ent/enttest"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func ok(n int) []models.SyllableView {
	out := make([]models.SyllableView, n)
	for i := range out {
		out[i].Status = "ok"
	}
	return out
}

func TestSpeakTier(t *testing.T) {
	warn := ok(2)
	warn[1].Status = "warn"
	cases := []struct {
		name       string
		assessable bool
		score      float64
		syllables  []models.SyllableView
		want       string
	}{
		{"not graded", false, 99, ok(2), None},
		{"low", true, 30, ok(2), Keep},
		{"close", true, 60, ok(2), Close},
		{"nice", true, 85, ok(2), Nice},
		{"perfect", true, 95, ok(2), Perfect},
		{"high score but a wobbly tone", true, 98, warn, Nice},
		{"high score, no syllables", true, 99, nil, Nice},
	}
	for _, c := range cases {
		if got := SpeakTier(c.assessable, c.score, c.syllables); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}

func TestMatchTierAndPoints(t *testing.T) {
	if got := MatchTier(false, 0, 0); got != None {
		t.Errorf("wrong pick = %q", got)
	}
	if got := MatchTier(true, 0, 100); got != Perfect || Points(got, models.ModePictureHanziMatch) != 6 {
		t.Errorf("first try = %q (%d xp)", got, Points(got, models.ModePictureHanziMatch))
	}
	if got := MatchTier(true, 1, 75); got != Close || Points(got, models.ModePictureHanziMatch) != 3 {
		t.Errorf("one miss = %q (%d xp)", got, Points(got, models.ModePictureHanziMatch))
	}
	if Points(Perfect, models.ModeImageSpeak) != 12 || Points(Keep, models.ModePictureHanziMatch) != 1 {
		t.Error("points table")
	}
}

// The rest runs against Postgres when TEST_DATABASE_URL is set (a database of its own: other
// packages migrate theirs in parallel).
type fixture struct {
	db   *ent.Client
	user *ent.User
	tz   *time.Location
}

func setup(t *testing.T) fixture {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db := enttest.Open(t, "postgres", url)
	u, err := db.User.Create().SetEmail(uuid.NewString() + "@example.com").SetPasswordHash("x").SetPasswordSalt("x").
		SetTimezone("America/New_York").Save(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	tz, _ := time.LoadLocation("America/New_York")
	return fixture{db, u, tz}
}

// grade saves a log for a new card with this word and awards it.
func (f fixture) grade(t *testing.T, word string, mode cardattemptlog.Mode, tier string, at time.Time) Result {
	t.Helper()
	ctx := context.Background()
	card := f.db.Flashcard.Create().SetOwnerID(f.user.ID).SetHanzi(word).SaveX(ctx)
	attempt := f.db.LessonAttempt.Create().SetOwnerID(f.user.ID).SetMode("hanzi_speak").SaveX(ctx)
	log := f.db.CardAttemptLog.Create().SetOwnerID(f.user.ID).SetUserID(f.user.ID).SetLessonAttemptID(attempt.ID).
		SetFlashcardID(card.ID).SetMode(mode).SetCreatedAt(at).SaveX(ctx)
	res, err := Award(ctx, f.db, f.user.ID, log, word, tier, 0)
	if err != nil {
		t.Fatal(err)
	}
	if stamped := f.db.CardAttemptLog.GetX(ctx, log.ID); stamped.XpEarned != res.Earned || stamped.XpTier != tier {
		t.Fatalf("log stamped %d/%q, want %d/%q", stamped.XpEarned, stamped.XpTier, res.Earned, tier)
	}
	return res
}

func (f fixture) total(t *testing.T) int {
	return f.db.User.GetX(context.Background(), f.user.ID).XpTotal
}

func TestTopUpPerWordPerDay(t *testing.T) {
	f := setup(t)
	speak := cardattemptlog.ModeHanziSpeak
	at := func(d, h int) time.Time { return time.Date(2026, 11, d, h, 0, 0, 0, f.tz) }

	steps := []struct {
		tier string
		at   time.Time
		want int
	}{
		{Close, at(1, 9), 5},    // first take
		{Keep, at(1, 10), 0},    // worse retry: nothing
		{Nice, at(1, 11), 3},    // better retry tops up to 8
		{Nice, at(1, 12), 0},    // same tier again: nothing
		{Perfect, at(1, 23), 4}, // 23:00 New York is the 2nd in UTC, still the 1st here
		{Perfect, at(2, 9), 12}, // a new day starts fresh
		{None, at(2, 10), 0},    // ungraded take
	}
	for i, s := range steps {
		if got := f.grade(t, "老公", speak, s.tier, s.at); got.Earned != s.want {
			t.Fatalf("step %d (%s): earned %d, want %d", i, s.tier, got.Earned, s.want)
		}
	}
	if got := f.total(t); got != 24 {
		t.Fatalf("total = %d, want 24", got)
	}
	// Another word the same day earns on its own; match earns half.
	if got := f.grade(t, "老婆", cardattemptlog.ModePictureHanziMatch, Perfect, at(2, 11)); got.Earned != 6 {
		t.Fatalf("match perfect earned %d, want 6", got.Earned)
	}
	// Speaking the same word after matching it tops up from 6 to 12.
	if got := f.grade(t, "老婆", speak, Perfect, at(2, 12)); got.Earned != 6 {
		t.Fatalf("speak after match earned %d, want 6", got.Earned)
	}
	if got, _ := Today(context.Background(), f.db, f.db.User.GetX(context.Background(), f.user.ID), at(2, 13)); got != 24 {
		t.Fatalf("today = %d, want 24", got)
	}
}

func TestConcurrentGradesCountOnce(t *testing.T) {
	f := setup(t)
	at := time.Date(2026, 11, 5, 9, 0, 0, 0, f.tz)
	ctx := context.Background()
	attempt := f.db.LessonAttempt.Create().SetOwnerID(f.user.ID).SetMode("hanzi_speak").SaveX(ctx)
	card := f.db.Flashcard.Create().SetOwnerID(f.user.ID).SetHanzi("妈妈").SaveX(ctx)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		log := f.db.CardAttemptLog.Create().SetOwnerID(f.user.ID).SetUserID(f.user.ID).SetLessonAttemptID(attempt.ID).
			SetFlashcardID(card.ID).SetMode(cardattemptlog.ModeHanziSpeak).SetCreatedAt(at).SaveX(ctx)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := Award(ctx, f.db, f.user.ID, log, "妈妈", Nice, 90); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if got := f.total(t); got != 8 {
		t.Fatalf("total after 8 concurrent Nice grades = %d, want 8", got)
	}
}

func TestAwardOnce(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	at := func(d, h int) time.Time { return time.Date(2026, 11, d, h, 0, 0, 0, f.tz) }
	steps := []struct {
		tier string
		at   time.Time
		want int
	}{
		{Close, at(1, 9), 5},
		{Keep, at(1, 10), 0},
		{Nice, at(2, 9), 3},     // another day tops up, never resets
		{Nice, at(3, 9), 0},     // already paid 8
		{Perfect, at(3, 10), 4}, // tops up to 12, the most it ever pays
		{Perfect, at(4, 9), 0},
	}
	for i, s := range steps {
		res, err := AwardOnce(ctx, f.db, f.user.ID, "tones/tone-1/ma", s.tier, 0, s.at)
		if err != nil {
			t.Fatal(err)
		}
		if res.Earned != s.want {
			t.Fatalf("step %d: earned %d, want %d", i, res.Earned, s.want)
		}
	}
	if got := f.total(t); got != 12 {
		t.Fatalf("total %d, want 12", got)
	}
	u := f.db.User.GetX(ctx, f.user.ID)
	if got, _ := Today(ctx, f.db, u, at(3, 12)); got != 4 {
		t.Fatalf("XP today on day 3 = %d, want 4 (only what that day earned)", got)
	}
}
