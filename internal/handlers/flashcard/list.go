package flashcard

import (
	"context"
	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardattemptlog"
	"net/http"
	"sort"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	entcategory "github.com/csotherden/mandarin-flashcards/ent/category"
	entflashcard "github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/setflashcard"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"github.com/google/uuid"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	query := cardQuery(h.db).
		Where(entflashcard.HasOwnerWith(user.ID(userID))).
		Order(entflashcard.ByCreatedAt(sql.OrderDesc()))
	if term := r.URL.Query().Get("q"); term != "" {
		query = query.Where(entflashcard.Or(
			entflashcard.HanziContainsFold(term),
			entflashcard.PinyinContainsFold(term),
			entflashcard.EnglishContainsFold(term),
		))
	}
	if raw := r.URL.Query().Get("category_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Invalid category id", err)
			return
		}
		query = query.Where(entflashcard.HasCategoriesWith(entcategory.ID(id)))
	}
	if raw := r.URL.Query().Get("set_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Invalid set id", err)
			return
		}
		query = query.Where(entflashcard.HasSetMembershipsWith(setflashcard.SetID(id)))
	}
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "oldest" {
		query = query.Order(entflashcard.ByCreatedAt(sql.OrderAsc()))
	}
	rows, err := query.All(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load flashcards", err)
		return
	}
	scores, err := recentScores(r.Context(), h.db, userID)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load flashcards", err)
		return
	}
	out := make([]models.FlashcardResponse, 0, len(rows))
	for _, row := range rows {
		card := present.Flashcard(row)
		if s, ok := scores[row.ID]; ok {
			card.RecentScore = &s
		}
		out = append(out, card)
	}
	SortCards(out, sortBy)
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, out)
}

// recentWindow is how far back a card's recent score looks.
const recentWindow = 30 * 24 * time.Hour

// recentScores averages each card's grades over the last 30 days.
func recentScores(ctx context.Context, db *ent.Client, userID uuid.UUID) (map[uuid.UUID]float64, error) {
	logs, err := db.CardAttemptLog.Query().
		Where(cardattemptlog.UserID(userID), cardattemptlog.CreatedAtGT(time.Now().Add(-recentWindow))).
		Select(cardattemptlog.FieldFlashcardID, cardattemptlog.FieldOverallScore).
		All(ctx)
	if err != nil {
		return nil, err
	}
	sums := map[uuid.UUID]float64{}
	counts := map[uuid.UUID]int{}
	for _, l := range logs {
		sums[l.FlashcardID] += l.OverallScore
		counts[l.FlashcardID]++
	}
	out := make(map[uuid.UUID]float64, len(sums))
	for id, sum := range sums {
		out[id] = sum / float64(counts[id])
	}
	return out, nil
}

// SortCards orders a card list in place. "az" sorts by pinyin (then hanzi); "weakest" puts the
// lowest recent score first and cards not practiced lately last; anything else keeps the
// query's order (newest or oldest first).
func SortCards(cards []models.FlashcardResponse, by string) {
	switch by {
	case "az":
		sort.SliceStable(cards, func(i, j int) bool {
			a, b := pinyinKey(cards[i]), pinyinKey(cards[j])
			if a != b {
				return a < b
			}
			return cards[i].Hanzi < cards[j].Hanzi
		})
	case "weakest":
		sort.SliceStable(cards, func(i, j int) bool {
			a, b := cards[i].RecentScore, cards[j].RecentScore
			switch {
			case a == nil:
				return false
			case b == nil:
				return true
			default:
				return *a < *b
			}
		})
	}
}

// pinyinKey is the pinyin with tone marks folded away, for alphabetical order.
func pinyinKey(c models.FlashcardResponse) string {
	key := strings.ToLower(c.TonalPinyin)
	if key == "" {
		key = strings.ToLower(c.Pinyin)
	}
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r == ' ' || r == '\'' {
			return -1
		}
		return r
	}, key)
}
