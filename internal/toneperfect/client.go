package toneperfect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/google/uuid"
)

const defaultURL = "https://api.toneperfect.app/v1/assess"

// ErrNotConfigured means TONEPERFECT_API_KEY is unset.
var ErrNotConfigured = errors.New("toneperfect is not configured")

// Result is the normalized grade plus the raw provider JSON.
type Result struct {
	Assessable   bool
	Reason       string
	Overall      float64
	ToneAccuracy *float64
	Syllables    []models.SyllableView
	Provider     map[string]any
}

type Client struct {
	apiKey string
	url    string
	http   *http.Client
}

func New(apiKey, rawURL string) *Client {
	apiKey = strings.TrimSpace(apiKey)
	if rawURL = strings.TrimSpace(rawURL); rawURL == "" {
		rawURL = defaultURL
	}
	return &Client{
		apiKey: apiKey,
		url:    rawURL,
		http:   &http.Client{Timeout: 45 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.apiKey != ""
}

// Assess posts 16 kHz mono WAV and the expected hanzi. The compact response
// schema is requested so per-character initial, final, and tone stay stable.
func (c *Client) Assess(ctx context.Context, wav []byte, hanzi, pinyin string) (*Result, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("audio", "take.wav")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(wav); err != nil {
		return nil, err
	}
	if err := writer.WriteField("text", hanzi); err != nil {
		return nil, err
	}
	if err := writer.WriteField("language", "zh-CN"); err != nil {
		return nil, err
	}
	if pinyin = requestPinyin(hanzi, pinyin); pinyin != "" {
		if err := writer.WriteField("pinyin", pinyin); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("TonePerfect-Response", "compact")
	req.Header.Set("Idempotency-Key", uuid.NewString())

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("toneperfect: %w", err)
	}
	defer res.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, statusError(res.StatusCode, payload)
	}

	parsed, err := ParseCompact(payload)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

// statusError names TonePerfect's error code and message when the body has them
// ({"error": {"code", "message"}, "request_id"}), so a refusal is diagnosable from the log.
func statusError(status int, payload []byte) error {
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		RequestID string `json:"request_id"`
	}
	if json.Unmarshal(payload, &body) != nil || body.Error.Code == "" {
		return fmt.Errorf("toneperfect status %d", status)
	}
	return fmt.Errorf("toneperfect status %d %s: %s (request %s)", status, body.Error.Code, body.Error.Message, body.RequestID)
}

func ParseCompact(payload []byte) (*Result, error) {
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("toneperfect response: %w", err)
	}
	var doc compactResponse
	if err := json.Unmarshal(payload, &doc); err != nil {
		return nil, fmt.Errorf("toneperfect response: %w", err)
	}
	result := &Result{
		Assessable: doc.Assessable,
		Reason:     doc.Reason,
		Provider:   raw,
		Syllables:  syllablesFrom(doc),
	}
	if doc.Scores != nil {
		if doc.Scores.Overall != nil {
			result.Overall = *doc.Scores.Overall
		}
		result.ToneAccuracy = doc.Scores.Tone
	}
	if !doc.Assessable && result.Reason == "" {
		result.Reason = "This recording could not be graded"
	}
	return result, nil
}
