package present

import (
	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/google/uuid"
)

func User(u *ent.User) models.UserResponse {
	return models.UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		Preferences: u.Preferences,
		Timezone:    u.Timezone,
		CreatedAt:   u.CreatedAt,
	}
}

func Category(c *ent.Category) models.CategoryResponse {
	return models.CategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Color:     c.Color,
		CreatedAt: c.CreatedAt,
	}
}

func Flashcard(f *ent.Flashcard) models.FlashcardResponse {
	resp := models.FlashcardResponse{
		ID:                    f.ID,
		Hanzi:                 f.Hanzi,
		Pinyin:                f.Pinyin,
		TonalPinyin:           f.TonalPinyin,
		English:               f.English,
		ExampleSentenceHanzi:  f.ExampleSentenceHanzi,
		ExampleSentencePinyin: f.ExampleSentencePinyin,
		ImageURL:              f.ImageURL,
		HasLocalImage:         f.ImageObjectKey != nil && *f.ImageObjectKey != "",
		UnsplashAttribution:   f.UnsplashAttribution,
		CategoryIDs:           []uuid.UUID{},
		SetIDs:                []uuid.UUID{},
		CreatedAt:             f.CreatedAt,
	}
	for _, category := range f.Edges.Categories {
		resp.CategoryIDs = append(resp.CategoryIDs, category.ID)
	}
	for _, membership := range f.Edges.SetMemberships {
		resp.SetIDs = append(resp.SetIDs, membership.SetID)
	}
	return resp
}

func Set(s *ent.CardSet, cardCount int) models.SetResponse {
	return models.SetResponse{
		ID:          s.ID,
		Title:       s.Title,
		Description: s.Description,
		CardCount:   cardCount,
		CreatedAt:   s.CreatedAt,
	}
}

func Attempt(a *ent.LessonAttempt) models.AttemptResponse {
	ids := a.CardIds
	if ids == nil {
		ids = []uuid.UUID{}
	}
	return models.AttemptResponse{
		ID:                 a.ID,
		LessonID:           a.LessonID,
		Mode:               string(a.Mode),
		CardCountRequested: a.CardCountRequested,
		CardIDs:            ids,
		OverallScore:       a.OverallScore,
		StartedAt:          a.StartedAt,
		CompletedAt:        a.CompletedAt,
	}
}

func CardResult(log *ent.CardAttemptLog) models.CardResult {
	result := models.CardResult{
		FlashcardID:       log.FlashcardID,
		OverallScore:      log.OverallScore,
		ToneAccuracyScore: log.ToneAccuracyScore,
		CreatedAt:         log.CreatedAt,
	}
	if log.LlmCoachingTip != nil {
		result.CoachingTip = *log.LlmCoachingTip
	}
	if log.RawAPIFeedback != nil {
		result.Syllables = log.RawAPIFeedback.Syllables
	}
	return result
}
