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

// Timezone is the device's IANA zone; it sets which calendar day counts toward the streak.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`
}

type TimezoneRequest struct {
	Timezone string `json:"timezone"`
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
	Timezone    string      `json:"timezone"`
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
	TonalPinyin           string               `json:"tonal_pinyin"`
	English               string               `json:"english"`
	ExampleSentenceHanzi  string               `json:"example_sentence_hanzi"`
	ExampleSentencePinyin string               `json:"example_sentence_pinyin"`
	ImageURL              string               `json:"image_url"`
	UnsplashAttribution   *UnsplashAttribution `json:"unsplash_attribution"`
	CategoryIDs           []uuid.UUID          `json:"category_ids"`
	// CategoryNames are categories by name: existing ones (any case) are linked, missing ones made.
	CategoryNames []string    `json:"category_names"`
	SetIDs        []uuid.UUID `json:"set_ids"`
}

type FlashcardResponse struct {
	ID                    uuid.UUID            `json:"id"`
	Hanzi                 string               `json:"hanzi"`
	Pinyin                string               `json:"pinyin"`
	TonalPinyin           string               `json:"tonal_pinyin"`
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

// AutofillRequest is what the user typed: hanzi, English or pinyin.
type AutofillRequest struct {
	Input string `json:"input"`
}

type AutofillResponse struct {
	Hanzi                 string `json:"hanzi"`
	Pinyin                string `json:"pinyin"`
	TonalPinyin           string `json:"tonal_pinyin"`
	English               string `json:"english"`
	ExampleSentenceHanzi  string `json:"example_sentence_hanzi"`
	ExampleSentencePinyin string `json:"example_sentence_pinyin"`
	ImageSearch           string `json:"image_search"`
	// Categories are suggested topic categories, reusing the user's own names where one fits.
	Categories []string `json:"categories"`
	// ExistingCardID is the user's card with the same hanzi, when there is one.
	ExistingCardID *uuid.UUID `json:"existing_card_id,omitempty"`
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
	LessonID *uuid.UUID `json:"lesson_id"`
	SetID    *uuid.UUID `json:"set_id"`
	// SetIDs draws from several sets at once; All draws from every card the user has.
	SetIDs          []uuid.UUID `json:"set_ids"`
	All             bool        `json:"all"`
	SourceAttemptID *uuid.UUID  `json:"source_attempt_id"`
	Mode            string      `json:"mode"`
	CardCount       int         `json:"card_count"`
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
	// XPEarned is what this grade added (0 when the word already reached this tier today);
	// XPTier is the tier it reached (keep, close, nice, perfect, or empty when not graded).
	XPEarned  int       `json:"xp_earned"`
	XPTier    string    `json:"xp_tier"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionResponse struct {
	Attempt      AttemptResponse     `json:"attempt"`
	Cards        []FlashcardResponse `json:"cards"`
	Results      []CardResult        `json:"results"`
	OverallScore *float64            `json:"overall_score,omitempty"`
	// XPEarned counts every take in the session, retries included (Results keeps only the
	// latest take per card).
	XPEarned int `json:"xp_earned"`
}

type AssessResponse struct {
	Result     CardResult `json:"result"`
	Assessable bool       `json:"assessable"`
	Reason     string     `json:"reason,omitempty"`
	XPTotal    int        `json:"xp_total"`
}

type MatchInput struct {
	FlashcardID   uuid.UUID `json:"flashcard_id"`
	Correct       bool      `json:"correct"`
	MismatchCount int       `json:"mismatch_count"`
}

// HomeResponse is the Today screen.
type HomeResponse struct {
	// StreakDays is the live streak; StreakToday says whether today already counts toward it.
	StreakDays    int  `json:"streak_days"`
	StreakToday   bool `json:"streak_today"`
	StreakLongest int  `json:"streak_longest"`
	XPTotal       int  `json:"xp_total"`
	XPToday       int  `json:"xp_today"`
	// Intro to tones: lessons finished, lessons in all, and the next unfinished lesson's key
	// and title (empty once the tutorial is done).
	TonesCompleted int            `json:"tones_completed"`
	TonesTotal     int            `json:"tones_total"`
	TonesNext      string         `json:"tones_next"`
	TonesNextTitle string         `json:"tones_next_title"`
	PracticedToday int            `json:"practiced_today"`
	TotalCards     int            `json:"total_cards"`
	SuggestedCount int            `json:"suggested_count"`
	LastMode       string         `json:"last_mode,omitempty"`
	Sets           []HomeSet      `json:"sets"`
	Slipped        []SlippedCard  `json:"slipped"`
	SavedSessions  []SavedSession `json:"saved_sessions"`
	// OpenSession is the latest session left unfinished in the last week, to pick up again.
	OpenSession *OpenSession `json:"open_session,omitempty"`
}

// OpenSession is a started, unfinished session: Done of Total cards graded so far.
type OpenSession struct {
	ID        uuid.UUID `json:"id"`
	Mode      string    `json:"mode"`
	Done      int       `json:"done"`
	Total     int       `json:"total"`
	StartedAt time.Time `json:"started_at"`
}

// HomeSet is a set with a card whose picture can stand for it.
type HomeSet struct {
	Set   SetResponse        `json:"set"`
	Cover *FlashcardResponse `json:"cover,omitempty"`
}

// SlippedCard is a recently practiced card whose average score is under 85.
type SlippedCard struct {
	Card    FlashcardResponse `json:"card"`
	Average float64           `json:"average"`
}

// SavedSession is a saved lesson the Today screen can start in one tap.
type SavedSession struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	CardCount int       `json:"card_count"`
}

// GenerateSetRequest asks for a themed set made on autopilot; SetID adds to an existing set.
type GenerateSetRequest struct {
	Theme string     `json:"theme"`
	Count int        `json:"count"`
	SetID *uuid.UUID `json:"set_id"`
}

// GenerateSetResponse is the set and the cards made for it; Skipped are words left out
// (already known, or not fillable).
type GenerateSetResponse struct {
	Set     SetResponse         `json:"set"`
	Cards   []FlashcardResponse `json:"cards"`
	Skipped []string            `json:"skipped"`
}

// LibrarySetResponse is a set in the global library, as browsed. ImportedSetID is the
// viewer's own copy, when they have one.
type LibrarySetResponse struct {
	ID            uuid.UUID  `json:"id"`
	Slug          string     `json:"slug"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Level         string     `json:"level"`
	CardCount     int        `json:"card_count"`
	CoverURL      string     `json:"cover_url"`
	Preview       []string   `json:"preview"`
	ImportedSetID *uuid.UUID `json:"imported_set_id,omitempty"`
}

type LibraryCardResponse struct {
	ID                    uuid.UUID            `json:"id"`
	Hanzi                 string               `json:"hanzi"`
	Pinyin                string               `json:"pinyin"`
	TonalPinyin           string               `json:"tonal_pinyin"`
	English               string               `json:"english"`
	ExampleSentenceHanzi  string               `json:"example_sentence_hanzi"`
	ExampleSentencePinyin string               `json:"example_sentence_pinyin"`
	ImageURL              string               `json:"image_url"`
	UnsplashAttribution   *UnsplashAttribution `json:"unsplash_attribution,omitempty"`
	// Owned is true when the viewer already has this word (importing reuses their card).
	Owned bool `json:"owned"`
}

type LibrarySetDetail struct {
	LibrarySetResponse
	Cards []LibraryCardResponse `json:"cards"`
}

// LibraryImportResponse is the viewer's copy of a library set after importing it.
type LibraryImportResponse struct {
	Set      SetResponse `json:"set"`
	Created  bool        `json:"created"`
	NewCards int         `json:"new_cards"`
}
