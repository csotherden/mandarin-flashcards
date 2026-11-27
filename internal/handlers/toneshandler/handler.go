// Package toneshandler serves the "Intro to tones" tutorial: the lessons with the user's
// progress, grading a spoken item, and finishing a lesson.
package toneshandler

import (
	"errors"
	"net/http"
	"time"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/toneattempt"
	"github.com/csotherden/mandarin-flashcards/ent/toneprogress"
	"github.com/csotherden/mandarin-flashcards/internal/audio"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/ratelimit"
	"github.com/csotherden/mandarin-flashcards/internal/streak"
	"github.com/csotherden/mandarin-flashcards/internal/toneperfect"
	"github.com/csotherden/mandarin-flashcards/internal/tones"
	"github.com/csotherden/mandarin-flashcards/internal/xp"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	db     *ent.Client
	grader *toneperfect.Client
	limit  *ratelimit.Limiter
	logger *zap.Logger
}

func NewHandler(db *ent.Client, grader *toneperfect.Client, limit *ratelimit.Limiter, logger *zap.Logger) *Handler {
	return &Handler{db: db, grader: grader, limit: limit, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/tones", h.List).Methods(http.MethodGet)
	router.HandleFunc("/tones/assess", h.Assess).Methods(http.MethodPost)
	router.HandleFunc("/tones/lessons/{key}/complete", h.Complete).Methods(http.MethodPost)
}

// LessonView is a catalog lesson with the viewer's progress.
type LessonView struct {
	tones.Lesson
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	BestScore   float64    `json:"best_score"`
}

type ListResponse struct {
	Lessons   []LessonView `json:"lessons"`
	Completed int          `json:"completed"`
	// Next is the first unfinished lesson, empty once the tutorial is done.
	Next string `json:"next"`
}

// AssessResponse is one graded take of a tutorial item.
type AssessResponse struct {
	OverallScore float64               `json:"overall_score"`
	Assessable   bool                  `json:"assessable"`
	Reason       string                `json:"reason,omitempty"`
	Syllables    []models.SyllableView `json:"syllables"`
	XPEarned     int                   `json:"xp_earned"`
	XPTier       string                `json:"xp_tier"`
	// FirstPass is true while the lesson hasn't been finished: takes earn XP. After that the
	// lesson is practice, which still counts toward the streak.
	FirstPass bool `json:"first_pass"`
	XPTotal   int  `json:"xp_total"`
}

type CompleteResponse struct {
	FirstTime bool    `json:"first_time"`
	BestScore float64 `json:"best_score"`
	// XPEarned is everything this lesson's items have paid (only ever on the first pass).
	XPEarned int    `json:"xp_earned"`
	Next     string `json:"next"`
}

func (h *Handler) progress(r *http.Request, userID uuid.UUID) (map[string]*ent.ToneProgress, error) {
	rows, err := h.db.ToneProgress.Query().Where(toneprogress.UserID(userID)).All(r.Context())
	if err != nil {
		return nil, err
	}
	out := map[string]*ent.ToneProgress{}
	for _, row := range rows {
		out[row.LessonKey] = row
	}
	return out, nil
}

func next(done map[string]*ent.ToneProgress) string {
	for _, l := range tones.Lessons {
		if p := done[l.Key]; p == nil || p.CompletedAt == nil {
			return l.Key
		}
	}
	return ""
}

// List is GET /api/tones.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	done, err := h.progress(r, userID)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load the tones lessons", err)
		return
	}
	out := ListResponse{Lessons: make([]LessonView, 0, len(tones.Lessons)), Next: next(done)}
	for _, l := range tones.Lessons {
		v := LessonView{Lesson: l}
		if p := done[l.Key]; p != nil && p.CompletedAt != nil {
			v.Completed, v.CompletedAt, v.BestScore = true, p.CompletedAt, p.BestScore
			out.Completed++
		}
		out.Lessons = append(out.Lessons, v)
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, out)
}

func xpKey(lesson, item string) string { return "tones/" + lesson + "/" + item }

// Assess is POST /api/tones/assess, multipart: audio, lesson, item.
func (h *Handler) Assess(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	if !h.limit.Allow(userID.String()) {
		handlerutils.RespondWithError(w, h.logger, http.StatusTooManyRequests, handlerutils.CodeNotSet, "Too many pronunciation checks. Wait a minute and try again", nil)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, audio.MaxUpload+4096)
	if err := r.ParseMultipartForm(audio.MaxUpload); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "The recording is too large or invalid", err)
		return
	}
	lesson, item := tones.Find(r.FormValue("lesson"), r.FormValue("item"))
	if lesson == nil || item == nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Lesson item not found", nil)
		return
	}
	wav, err := audio.FromForm(r)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, err.Error(), err)
		return
	}
	grade, err := h.grader.Assess(r.Context(), wav, item.Hanzi, item.Tonal)
	if err != nil {
		if errors.Is(err, toneperfect.ErrNotConfigured) {
			handlerutils.RespondWithError(w, h.logger, http.StatusServiceUnavailable, handlerutils.CodeNotSet, "Pronunciation grading is not configured", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusBadGateway, handlerutils.CodeNotSet, "Pronunciation grading failed. Try again", err)
		return
	}
	ctx := r.Context()
	tier := xp.SpeakTier(grade.Assessable, grade.Overall, grade.Syllables)
	row, err := h.db.ToneAttempt.Create().
		SetUserID(userID).SetLessonKey(lesson.Key).SetItemKey(item.Key).
		SetOverallScore(grade.Overall).SetXpTier(tier).
		SetRawAPIFeedback(&models.StoredFeedback{Assessable: grade.Assessable, Reason: grade.Reason, Syllables: grade.Syllables, Provider: grade.Provider}).
		Save(ctx)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not save the grade", err)
		return
	}
	out := AssessResponse{
		OverallScore: grade.Overall, Assessable: grade.Assessable, Reason: grade.Reason,
		Syllables: grade.Syllables, XPTier: tier,
	}
	if out.Syllables == nil {
		out.Syllables = []models.SyllableView{}
	}
	// Every take counts toward the streak, first pass or practice, as card takes do.
	if err := streak.Record(ctx, h.db, userID, row.CreatedAt); err != nil {
		h.logger.Warn("Could not update the streak", zap.Error(err))
	}
	done, err := h.db.ToneProgress.Query().
		Where(toneprogress.UserID(userID), toneprogress.LessonKey(lesson.Key), toneprogress.CompletedAtNotNil()).
		Exist(ctx)
	if err != nil {
		h.logger.Warn("Could not read tones progress", zap.Error(err))
		done = true // never pay twice on a lookup failure
	}
	out.FirstPass = !done
	if out.FirstPass {
		res, err := xp.AwardOnce(ctx, h.db, userID, xpKey(lesson.Key, item.Key), tier, grade.Overall, row.CreatedAt)
		if err != nil {
			h.logger.Warn("Could not award XP", zap.Error(err))
		} else {
			out.XPEarned, out.XPTotal = res.Earned, res.Total
			if res.Earned > 0 {
				if err := row.Update().SetXpEarned(res.Earned).Exec(ctx); err != nil {
					h.logger.Warn("Could not stamp tones XP", zap.Error(err))
				}
			}
		}
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, out)
}

// Complete is POST /api/tones/lessons/{key}/complete. The first call marks the lesson done (its
// XP is closed from then on); later calls just refresh the best score.
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	lesson, _ := tones.Find(mux.Vars(r)["key"], "")
	if lesson == nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Lesson not found", nil)
		return
	}
	ctx := r.Context()
	fail := func(err error) {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not finish the lesson", err)
	}
	takes, err := h.db.ToneAttempt.Query().
		Where(toneattempt.UserID(userID), toneattempt.LessonKey(lesson.Key)).
		All(ctx)
	if err != nil {
		fail(err)
		return
	}
	best := map[string]float64{}
	earned := 0
	for _, t := range takes {
		best[t.ItemKey] = max(best[t.ItemKey], t.OverallScore)
		earned += t.XpEarned
	}
	sum := 0.0
	for _, it := range lesson.Items {
		sum += best[it.Key] // unattempted items count as 0
	}
	score := sum / float64(len(lesson.Items))

	out := CompleteResponse{BestScore: score, XPEarned: earned}
	now := time.Now()
	prior, err := h.db.ToneProgress.Query().Where(toneprogress.UserID(userID), toneprogress.LessonKey(lesson.Key)).Only(ctx)
	switch {
	case ent.IsNotFound(err):
		err = h.db.ToneProgress.Create().SetUserID(userID).SetLessonKey(lesson.Key).SetCompletedAt(now).SetBestScore(score).Exec(ctx)
		out.FirstTime = err == nil
	case err != nil:
	default:
		update := prior.Update().SetBestScore(score)
		if prior.CompletedAt == nil {
			update.SetCompletedAt(now)
			out.FirstTime = true
		}
		err = update.Exec(ctx)
	}
	if err != nil && !ent.IsConstraintError(err) {
		fail(err)
		return
	}
	done, err := h.progress(r, userID)
	if err != nil {
		fail(err)
		return
	}
	out.Next = next(done)
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, out)
}
