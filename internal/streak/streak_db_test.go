package streak

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardattemptlog"
	"github.com/csotherden/mandarin-flashcards/ent/enttest"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Runs against a real Postgres when TEST_DATABASE_URL is set, e.g.
// TEST_DATABASE_URL="postgres://a:a@localhost:5432/streaktest?sslmode=disable" go test ./internal/streak/
// go test runs packages in parallel and each migrates its database, so give each package its own.
func testDB(t *testing.T) *ent.Client {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	return enttest.Open(t, "postgres", url)
}

func newUser(t *testing.T, db *ent.Client, tz string) *ent.User {
	t.Helper()
	u, err := db.User.Create().
		SetEmail(uuid.NewString() + "@example.com").
		SetPasswordHash("x").
		SetPasswordSalt("x").
		SetTimezone(tz).
		Save(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// grade stores a graded card at the given time, the way assess and match do.
func grade(t *testing.T, db *ent.Client, u *ent.User, at time.Time) {
	t.Helper()
	ctx := context.Background()
	card, err := db.Flashcard.Create().SetOwnerID(u.ID).SetHanzi("好").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	attempt, err := db.LessonAttempt.Create().SetOwnerID(u.ID).SetMode("hanzi_speak").SetCardCountRequested(1).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.CardAttemptLog.Create().
		SetOwnerID(u.ID).SetUserID(u.ID).SetLessonAttemptID(attempt.ID).SetFlashcardID(card.ID).
		SetMode(cardattemptlog.ModeHanziSpeak).SetOverallScore(90).
		SetCreatedAt(at).
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := Record(ctx, db, u.ID, at); err != nil {
		t.Fatal(err)
	}
}

func check(t *testing.T, db *ent.Client, u *ent.User, now time.Time) Status {
	t.Helper()
	s, err := Check(context.Background(), db, u.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestRecordAndCheck(t *testing.T) {
	db := testDB(t)
	ny, _ := time.LoadLocation("America/New_York")
	u := newUser(t, db, "America/New_York")
	day := func(d, h int) time.Time { return time.Date(2026, 11, d, h, 0, 0, 0, ny) }

	if s := check(t, db, u, day(1, 9)); s != (Status{}) {
		t.Fatalf("no practice: %+v", s)
	}
	grade(t, db, u, day(1, 9))
	grade(t, db, u, day(1, 20)) // same day twice counts once
	if s := check(t, db, u, day(1, 21)); s != (Status{Current: 1, Longest: 1, Today: true}) {
		t.Fatalf("after day 1: %+v", s)
	}
	// Next day before practicing: alive, not done today.
	if s := check(t, db, u, day(2, 8)); s != (Status{Current: 1, Longest: 1}) {
		t.Fatalf("day 2 morning: %+v", s)
	}
	grade(t, db, u, day(2, 8))
	grade(t, db, u, day(3, 23)) // 23:00 New York is already the 4th in UTC; still the 3rd here
	if s := check(t, db, u, day(3, 23)); s != (Status{Current: 3, Longest: 3, Today: true}) {
		t.Fatalf("day 3: %+v", s)
	}
	// A missed day: the read settles it to 0 and stores that.
	if s := check(t, db, u, day(5, 9)); s != (Status{Current: 0, Longest: 3}) {
		t.Fatalf("after a gap: %+v", s)
	}
	if got, _ := db.User.Get(context.Background(), u.ID); got.StreakCurrent != 0 {
		t.Fatalf("lapse not stored: %d", got.StreakCurrent)
	}
	grade(t, db, u, day(5, 10))
	if s := check(t, db, u, day(5, 11)); s != (Status{Current: 1, Longest: 3, Today: true}) {
		t.Fatalf("restart: %+v", s)
	}
}

func TestBackfillHasNoWindow(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	u := newUser(t, db, "UTC")
	start := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	// 200 days of history written directly, as it existed before streaks were stored.
	card, _ := db.Flashcard.Create().SetOwnerID(u.ID).SetHanzi("好").Save(ctx)
	attempt, _ := db.LessonAttempt.Create().SetOwnerID(u.ID).SetMode("hanzi_speak").SetCardCountRequested(1).Save(ctx)
	for i := 0; i < 200; i++ {
		_, err := db.CardAttemptLog.Create().
			SetOwnerID(u.ID).SetUserID(u.ID).SetLessonAttemptID(attempt.ID).SetFlashcardID(card.ID).
			SetMode(cardattemptlog.ModeHanziSpeak).SetCreatedAt(start.AddDate(0, 0, i)).
			Save(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
	now := start.AddDate(0, 0, 200) // the day after the last practice
	if s := check(t, db, u, now); s != (Status{Current: 200, Longest: 200}) {
		t.Fatalf("backfill: %+v", s)
	}
}
