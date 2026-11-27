package autofill

import (
	"errors"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/csotherden/mandarin-flashcards/ent/category"
	"github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/llm"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/pinyin"
	"go.uber.org/zap"
)

// Autofill suggests a card from hanzi, an English word or pinyin.
func (h *Handler) Autofill(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	if !h.limit.Allow(userID.String()) {
		handlerutils.RespondWithError(w, h.logger, http.StatusTooManyRequests, handlerutils.CodeNotSet, "Too many suggestions. Wait a minute and try again", nil)
		return
	}
	var req models.AutofillRequest
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	input := strings.TrimSpace(req.Input)
	if input == "" {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Enter hanzi, English or pinyin", nil)
		return
	}
	if utf8.RuneCountInString(input) > 100 {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Enter a word or a short phrase", nil)
		return
	}
	existing, err := h.db.Category.Query().Where(category.HasOwnerWith(user.ID(userID))).Select(category.FieldName).Strings(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not suggest card details", err)
		return
	}
	filled, err := h.llm.Autofill(r.Context(), input, existing)
	if err != nil {
		if errors.Is(err, llm.ErrNotConfigured) {
			handlerutils.RespondWithError(w, h.logger, http.StatusServiceUnavailable, handlerutils.CodeNotSet, "Card suggestions are not configured", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusBadGateway, handlerutils.CodeNotSet, "Could not suggest card details", err)
		return
	}
	// Hanzi the user typed stays exactly as typed.
	if isHanzi(input) {
		filled.Hanzi = input
	}
	// The numbered pinyin goes to TonePerfect as is: keep it only in the form it accepts.
	filled.TonalPinyin = pinyin.NormalizeTonal(filled.TonalPinyin)
	if filled.TonalPinyin != "" && !pinyin.ValidTonal(filled.Hanzi, filled.TonalPinyin) {
		h.logger.Warn("autofill: dropping tonal pinyin that is not one numbered syllable per character",
			zap.String("hanzi", filled.Hanzi), zap.String("tonal_pinyin", filled.TonalPinyin))
		filled.TonalPinyin = ""
	}
	filled.Categories = matchCategories(filled.Categories, existing)
	if card, err := h.db.Flashcard.Query().
		Where(flashcard.Hanzi(filled.Hanzi), flashcard.HasOwnerWith(user.ID(userID))).
		First(r.Context()); err == nil {
		filled.ExistingCardID = &card.ID
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, filled)
}

// matchCategories keeps at most three suggestions, spelled as the user's existing category
// when one matches (any case), without repeats.
func matchCategories(suggested, existing []string) []string {
	byKey := make(map[string]string, len(existing))
	for _, name := range existing {
		byKey[strings.ToLower(name)] = name
	}
	out := []string{}
	seen := map[string]bool{}
	for _, name := range suggested {
		name = strings.TrimSpace(name)
		key := strings.ToLower(name)
		if name == "" || seen[key] || utf8.RuneCountInString(name) > 40 {
			continue
		}
		seen[key] = true
		if own, ok := byKey[key]; ok {
			name = own
		}
		out = append(out, name)
		if len(out) == 3 {
			break
		}
	}
	return out
}

// isHanzi reports whether s is Chinese characters (and punctuation or spaces), with no Latin
// letters: what the user typed is the card's hanzi.
func isHanzi(s string) bool {
	han := false
	for _, r := range s {
		switch {
		case unicode.Is(unicode.Han, r):
			han = true
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			return false
		}
	}
	return han
}
