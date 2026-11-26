package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/csotherden/mandarin-flashcards/internal/models"
)

// ErrNotConfigured means LLM_API_KEY is unset.
var ErrNotConfigured = errors.New("llm is not configured")

type Client struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
}

func New(apiKey, baseURL, model string) *Client {
	apiKey = strings.TrimSpace(apiKey)
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.apiKey != ""
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *Client) complete(ctx context.Context, system, user string) (string, error) {
	if !c.Enabled() {
		return "", ErrNotConfigured
	}
	payload, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("llm status %d", res.StatusCode)
	}
	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

// Autofill asks the model for card fields and a concrete Unsplash search phrase.
func (c *Client) Autofill(ctx context.Context, hanzi, english string) (*models.AutofillResponse, error) {
	system := `You fill Mandarin flashcards. Reply with one JSON object and no markdown.
Keys: hanzi, pinyin (tone marks), english, example_sentence_hanzi, example_sentence_pinyin, image_search.
image_search is a short concrete English photo query. For abstract words pick a physical scene a photo can show, such as "convenience store" for 方便.
Keep the example sentence short and everyday.`
	user := fmt.Sprintf("hanzi: %s\nenglish: %s", strings.TrimSpace(hanzi), strings.TrimSpace(english))
	text, err := c.complete(ctx, system, user)
	if err != nil {
		return nil, err
	}
	var out models.AutofillResponse
	if err := json.Unmarshal([]byte(extractJSON(text)), &out); err != nil {
		return nil, fmt.Errorf("llm autofill json: %w", err)
	}
	out.Hanzi = strings.TrimSpace(out.Hanzi)
	out.Pinyin = strings.TrimSpace(out.Pinyin)
	out.English = strings.TrimSpace(out.English)
	out.ExampleSentenceHanzi = strings.TrimSpace(out.ExampleSentenceHanzi)
	out.ExampleSentencePinyin = strings.TrimSpace(out.ExampleSentencePinyin)
	out.ImageSearch = strings.TrimSpace(out.ImageSearch)
	if out.Hanzi == "" {
		return nil, fmt.Errorf("llm autofill omitted hanzi")
	}
	return &out, nil
}

// Coach turns a TonePerfect breakdown into one or two sentences of advice.
func (c *Client) Coach(ctx context.Context, hanzi string, overall float64, syllables []models.SyllableView) (string, error) {
	blob, err := json.Marshal(syllables)
	if err != nil {
		return "", err
	}
	system := `You coach Mandarin pronunciation. Reply with one or two sentences, no preamble.
Name the character, the expected tone, and what was heard. Be specific and calm.
Do not mention scores, JSON, or that you are a model.`
	user := fmt.Sprintf("word: %s\noverall: %.1f\nsyllables: %s", hanzi, overall, blob)
	text, err := c.complete(ctx, system, user)
	if err != nil {
		return "", err
	}
	text = strings.TrimSpace(text)
	if len([]rune(text)) > 400 {
		text = string([]rune(text)[:400])
	}
	return text, nil
}

func extractJSON(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return text
}
