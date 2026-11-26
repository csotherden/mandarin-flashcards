package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ModeImageSpeak        = "image_speak"
	ModeHanziSpeak        = "hanzi_speak"
	ModePictureHanziMatch = "picture_hanzi_match"
	CoachingThreshold     = 85.0
)

type Preferences struct {
	DefaultTTSSpeed float64 `json:"default_tts_speed"`
	PreferredVoice  string  `json:"preferred_voice"`
}

type UnsplashAttribution struct {
	PhotographerName string `json:"photographer_name"`
	PhotographerURL  string `json:"photographer_url"`
	PhotoURL         string `json:"photo_url"`
	DownloadLocation string `json:"download_location,omitempty"`
}

// StoredFeedback is what card_attempt_logs.raw_api_feedback keeps: the normalized
// view the UI renders, plus the provider payload for later analytics.
type StoredFeedback struct {
	Assessable bool           `json:"assessable,omitempty"`
	Reason     string         `json:"reason,omitempty"`
	Syllables  []SyllableView `json:"syllables,omitempty"`
	Provider   map[string]any `json:"provider,omitempty"`
	Match      *MatchFeedback `json:"match,omitempty"`
}

type MatchFeedback struct {
	Correct       bool `json:"correct"`
	MismatchCount int  `json:"mismatch_count"`
}

type SyllableView struct {
	Character    string  `json:"character"`
	Pinyin       string  `json:"pinyin"`
	Score        float64 `json:"score"`
	ExpectedTone string  `json:"expected_tone"`
	HeardTone    string  `json:"heard_tone"`
	Status       string  `json:"status"`
	Detail       string  `json:"detail,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func GenericErrorCode(status int) string {
	switch status {
	case 400:
		return "bad_request"
	case 401:
		return "unauthorized"
	case 403:
		return "forbidden"
	case 404:
		return "not_found"
	case 409:
		return "conflict"
	case 429:
		return "rate_limited"
	case 503:
		return "not_configured"
	default:
		if status >= 500 {
			return "internal_error"
		}
		return "error"
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID          uuid.UUID   `json:"id"`
	Email       string      `json:"email"`
	Preferences Preferences `json:"preferences"`
	CreatedAt   time.Time   `json:"created_at"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type CategoryResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

type CategoryInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type FlashcardInput struct {
	Hanzi                 string               `json:"hanzi"`
	Pinyin                string               `json:"pinyin"`
	English               string               `json:"english"`
	ExampleSentenceHanzi  string               `json:"example_sentence_hanzi"`
	ExampleSentencePinyin string               `json:"example_sentence_pinyin"`
	ImageURL              string               `json:"image_url"`
	UnsplashAttribution   *UnsplashAttribution `json:"unsplash_attribution"`
	CategoryIDs           []uuid.UUID          `json:"category_ids"`
	SetIDs                []uuid.UUID          `json:"set_ids"`
}

type FlashcardResponse struct {
	ID                    uuid.UUID            `json:"id"`
	Hanzi                 string               `json:"hanzi"`
	Pinyin                string               `json:"pinyin"`
	English               string               `json:"english"`
	ExampleSentenceHanzi  string               `json:"example_sentence_hanzi"`
	ExampleSentencePinyin string               `json:"example_sentence_pinyin"`
	ImageURL              string               `json:"image_url"`
	HasLocalImage         bool                 `json:"has_local_image"`
	UnsplashAttribution   *UnsplashAttribution `json:"unsplash_attribution,omitempty"`
	CategoryIDs           []uuid.UUID          `json:"category_ids"`
	SetIDs                []uuid.UUID          `json:"set_ids"`
	CreatedAt             time.Time            `json:"created_at"`
}

type AutofillRequest struct {
	Hanzi   string `json:"hanzi"`
	English string `json:"english"`
}

type AutofillResponse struct {
	Hanzi                 string `json:"hanzi"`
	Pinyin                string `json:"pinyin"`
	English               string `json:"english"`
	ExampleSentenceHanzi  string `json:"example_sentence_hanzi"`
	ExampleSentencePinyin string `json:"example_sentence_pinyin"`
	ImageSearch           string `json:"image_search"`
}

type ImageCandidate struct {
	ID               string `json:"id"`
	ThumbURL         string `json:"thumb_url"`
	FullURL          string `json:"full_url"`
	PhotographerName string `json:"photographer_name"`
	PhotographerURL  string `json:"photographer_url"`
	PhotoURL         string `json:"photo_url"`
	DownloadLocation string `json:"download_location"`
}

type SetInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type SetResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CardCount   int       `json:"card_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type SetDetail struct {
	SetResponse
	Cards []FlashcardResponse `json:"cards"`
}

type SetCardsInput struct {
	FlashcardIDs []uuid.UUID `json:"flashcard_ids"`
}

type LessonInput struct {
	Title  string      `json:"title"`
	SetIDs []uuid.UUID `json:"set_ids"`
}

type LessonResponse struct {
	ID           uuid.UUID   `json:"id"`
	Title        string      `json:"title"`
	SetIDs       []uuid.UUID `json:"set_ids"`
	CardCount    int         `json:"card_count"`
	EnabledCount int         `json:"enabled_count"`
	CreatedAt    time.Time   `json:"created_at"`
}

type LessonCard struct {
	Card    FlashcardResponse `json:"card"`
	Enabled bool              `json:"enabled"`
}

type LessonDetail struct {
	LessonResponse
	Sets  []SetResponse `json:"sets"`
	Cards []LessonCard  `json:"cards"`
}

type IDsInput struct {
	SetIDs       []uuid.UUID `json:"set_ids"`
	FlashcardIDs []uuid.UUID `json:"flashcard_ids"`
}

type StartSessionRequest struct {
	LessonID        *uuid.UUID `json:"lesson_id"`
	SetID           *uuid.UUID `json:"set_id"`
	SourceAttemptID *uuid.UUID `json:"source_attempt_id"`
	Mode            string     `json:"mode"`
	CardCount       int        `json:"card_count"`
}

type AttemptResponse struct {
	ID                 uuid.UUID   `json:"id"`
	LessonID           *uuid.UUID  `json:"lesson_id,omitempty"`
	Mode               string      `json:"mode"`
	CardCountRequested int         `json:"card_count_requested"`
	CardIDs            []uuid.UUID `json:"card_ids"`
	OverallScore       *float64    `json:"overall_score,omitempty"`
	StartedAt          time.Time   `json:"started_at"`
	CompletedAt        *time.Time  `json:"completed_at,omitempty"`
}

type CardResult struct {
	FlashcardID       uuid.UUID      `json:"flashcard_id"`
	OverallScore      float64        `json:"overall_score"`
	ToneAccuracyScore *float64       `json:"tone_accuracy_score,omitempty"`
	CoachingTip       string         `json:"coaching_tip,omitempty"`
	Syllables         []SyllableView `json:"syllables,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
}

type SessionResponse struct {
	Attempt      AttemptResponse     `json:"attempt"`
	Cards        []FlashcardResponse `json:"cards"`
	Results      []CardResult        `json:"results"`
	OverallScore *float64            `json:"overall_score,omitempty"`
}

type AssessResponse struct {
	Result     CardResult `json:"result"`
	Assessable bool       `json:"assessable"`
	Reason     string     `json:"reason,omitempty"`
}

type MatchInput struct {
	FlashcardID   uuid.UUID `json:"flashcard_id"`
	Correct       bool      `json:"correct"`
	MismatchCount int       `json:"mismatch_count"`
}
