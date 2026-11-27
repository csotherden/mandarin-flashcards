// Package home serves the Today screen: what to practice now, quick starts and the words that
// slipped, in one request.
package home

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardattemptlog"
	"github.com/csotherden/mandarin-flashcards/ent/cardset"
	"github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/lesson"
	"github.com/csotherden/mandarin-flashcards/ent/lessonattempt"
	"github.com/csotherden/mandarin-flashcards/ent/toneprogress"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"github.com/csotherden/mandarin-flashcards/internal/streak"
	deck "github.com/csotherden/mandarin-flashcards/internal/study"
	"github.com/csotherden/mandarin-flashcards/internal/tones"
	"github.com/csotherden/mandarin-flashcards/internal/xp"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// slipThreshold is the average score under which a recently practiced word counts as slipped.
const slipThreshold = 85

type Handler struct {
	db     *ent.Client
	logger *zap.Logger
}

func NewHandler(db *ent.Client, logger *zap.Logger) *Handler {
	return &Handler{db: db, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/home", h.Get).Methods(http.MethodGet)
}

// Get is GET /api/home. Days are the user's saved timezone's calendar days.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	ctx := r.Context()
	fail := func(err error) {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load today", err)
	}
	me, err := h.db.User.Get(ctx, userID)
	if err != nil {
		fail(err)
		return
	}
	loc := streak.Location(me)

	out := models.HomeResponse{Sets: []models.HomeSet{}, Slipped: []models.SlippedCard{}, SavedSessions: []models.SavedSession{}}
	total, err := h.db.Flashcard.Query().Where(flashcard.HasOwnerWith(user.ID(userID))).Count(ctx)
	if err != nil {
		fail(err)
		return
	}
	out.TotalCards = total

	status, err := streak.Check(ctx, h.db, userID, time.Now())
	if err != nil {
		fail(err)
		return
	}
	out.StreakDays = status.Current
	out.StreakLongest = status.Longest
	out.StreakToday = status.Today
	out.XPTotal = me.XpTotal
	finished, err := h.db.ToneProgress.Query().
		Where(toneprogress.UserID(userID), toneprogress.CompletedAtNotNil()).
		Select(toneprogress.FieldLessonKey).
		Strings(ctx)
	if err != nil {
		fail(err)
		return
	}
	done := map[string]bool{}
	for _, key := range finished {
		done[key] = true
	}
	out.TonesTotal = len(tones.Lessons)
	for _, l := range tones.Lessons {
		if done[l.Key] {
			out.TonesCompleted++
		} else if out.TonesNext == "" {
			out.TonesNext, out.TonesNextTitle = l.Key, l.Title
		}
	}
	if out.XPToday, err = xp.Today(ctx, h.db, me, time.Now()); err != nil {
		fail(err)
		return
	}

	last, err := h.db.LessonAttempt.Query().
		Where(lessonattempt.HasOwnerWith(user.ID(userID)), lessonattempt.CompletedAtNotNil()).
		Order(ent.Desc(lessonattempt.FieldCompletedAt)).First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		fail(err)
		return
	}
	if last != nil {
		out.LastMode = string(last.Mode)
	}
	if out.OpenSession, err = openSession(ctx, h.db, userID); err != nil {
		fail(err)
		return
	}
	// Finishing a newer session since means the user moved on; don't nag about the old one.
	if out.OpenSession != nil && last != nil && last.CompletedAt != nil && out.OpenSession.StartedAt.Before(*last.CompletedAt) {
		out.OpenSession = nil
	}

	// Cards practiced today, and the averages behind "words that slipped".
	logs, err := h.db.CardAttemptLog.Query().
		Where(cardattemptlog.UserID(userID), cardattemptlog.CreatedAtGT(time.Now().AddDate(0, 0, -14))).
		All(ctx)
	if err != nil {
		fail(err)
		return
	}
	today := time.Now().In(loc).Format("2006-01-02")
	sums := map[uuid.UUID]float64{}
	counts := map[uuid.UUID]int{}
	practicedToday := map[uuid.UUID]bool{}
	for _, l := range logs {
		sums[l.FlashcardID] += l.OverallScore
		counts[l.FlashcardID]++
		if l.CreatedAt.In(loc).Format("2006-01-02") == today {
			practicedToday[l.FlashcardID] = true
		}
	}
	out.PracticedToday = len(practicedToday)
	type slip struct {
		id  uuid.UUID
		avg float64
	}
	var slips []slip
	for id, n := range counts {
		if avg := sums[id] / float64(n); avg < slipThreshold {
			slips = append(slips, slip{id, avg})
		}
	}
	sort.Slice(slips, func(i, j int) bool { return slips[i].avg < slips[j].avg })
	if len(slips) > 6 {
		slips = slips[:6]
	}
	if len(slips) > 0 {
		ids := make([]uuid.UUID, len(slips))
		for i, s := range slips {
			ids[i] = s.id
		}
		cards, err := deck.LoadOwned(ctx, h.db, userID, ids)
		if err != nil {
			fail(err)
			return
		}
		byID := map[uuid.UUID]*ent.Flashcard{}
		for _, c := range cards {
			byID[c.ID] = c
		}
		for _, s := range slips {
			if c, ok := byID[s.id]; ok {
				out.Slipped = append(out.Slipped, models.SlippedCard{Card: present.Flashcard(c), Average: s.avg})
			}
		}
	}

	// Sets with a cover card, most recently created first.
	sets, err := h.db.CardSet.Query().Where(cardset.HasOwnerWith(user.ID(userID))).
		Order(ent.Desc(cardset.FieldCreatedAt)).All(ctx)
	if err != nil {
		fail(err)
		return
	}
	for _, s := range sets {
		cards, err := deck.LoadSet(ctx, h.db, userID, s.ID)
		if err != nil {
			fail(err)
			return
		}
		item := models.HomeSet{Set: present.Set(s, len(cards))}
		for _, c := range cards {
			if c.ImageURL != "" || (c.ImageObjectKey != nil && *c.ImageObjectKey != "") {
				cover := present.Flashcard(c)
				item.Cover = &cover
				break
			}
		}
		out.Sets = append(out.Sets, item)
	}

	// Saved sessions (lessons), newest first.
	lessons, err := h.db.Lesson.Query().Where(lesson.HasOwnerWith(user.ID(userID))).
		Order(ent.Desc(lesson.FieldCreatedAt)).Limit(12).All(ctx)
	if err != nil {
		fail(err)
		return
	}
	for _, l := range lessons {
		loaded, err := deck.LoadLesson(ctx, h.db, userID, l.ID)
		if err != nil {
			fail(err)
			return
		}
		enabled := 0
		for _, c := range loaded.Cards {
			if _, off := loaded.Disabled[c.ID]; !off {
				enabled++
			}
		}
		out.SavedSessions = append(out.SavedSessions, models.SavedSession{ID: l.ID, Title: l.Title, CardCount: enabled})
	}

	out.SuggestedCount = min(10, total)
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, out)
}

// openWindow is how long an unfinished session is offered for resuming.
const openWindow = 7 * 24 * time.Hour

// openSession finds the most recent unfinished session from the last week that has at least one
// graded card, with its progress.
func openSession(ctx context.Context, db *ent.Client, userID uuid.UUID) (*models.OpenSession, error) {
	a, err := db.LessonAttempt.Query().
		Where(lessonattempt.HasOwnerWith(user.ID(userID)), lessonattempt.CompletedAtIsNil(),
			lessonattempt.StartedAtGT(time.Now().Add(-openWindow))).
		Order(ent.Desc(lessonattempt.FieldStartedAt)).
		First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(a.CardIds) == 0 {
		return nil, nil
	}
	graded, err := db.CardAttemptLog.Query().
		Where(cardattemptlog.LessonAttemptID(a.ID)).
		Unique(true).
		Select(cardattemptlog.FieldFlashcardID).
		Strings(ctx)
	if err != nil {
		return nil, err
	}
	if len(graded) == 0 {
		// Nothing graded yet: resuming would be the same as starting fresh, so don't offer it.
		return nil, nil
	}
	return &models.OpenSession{ID: a.ID, Mode: string(a.Mode), Done: len(graded), Total: len(a.CardIds), StartedAt: a.StartedAt}, nil
}
