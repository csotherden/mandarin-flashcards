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

// Autofill asks the model for card fields and a concrete Unsplash search phrase, from whatever
// the user typed: hanzi, an English word or phrase, or pinyin.
func (c *Client) Autofill(ctx context.Context, input string, categories []string) (*models.AutofillResponse, error) {
	user := "input: " + strings.TrimSpace(input)
	if len(categories) > 0 {
		user += "\nexisting categories: " + strings.Join(categories, ", ")
	}
	text, err := c.complete(ctx, autofillPrompt, user)
	if err != nil {
		return nil, err
	}
	var out models.AutofillResponse
	if err := json.Unmarshal([]byte(extractJSON(text)), &out); err != nil {
		return nil, fmt.Errorf("llm autofill json: %w", err)
	}
	out.Hanzi = strings.TrimSpace(out.Hanzi)
	out.Pinyin = strings.TrimSpace(out.Pinyin)
	out.TonalPinyin = strings.TrimSpace(out.TonalPinyin)
	out.English = strings.TrimSpace(out.English)
	out.ExampleSentenceHanzi = strings.TrimSpace(out.ExampleSentenceHanzi)
	out.ExampleSentencePinyin = strings.TrimSpace(out.ExampleSentencePinyin)
	out.ImageSearch = strings.TrimSpace(out.ImageSearch)
	if out.Hanzi == "" {
		return nil, fmt.Errorf("llm autofill omitted hanzi")
	}
	return &out, nil
}

// autofillPrompt fills a card from hanzi, English or pinyin. tonal_pinyin is the form TonePerfect
// accepts for choosing readings (one numbered syllable per character); the handler drops it when
// the model gets that form wrong.
const autofillPrompt = `You fill Mandarin flashcards for a learner. The input is Chinese characters (hanzi), an English word or phrase, or pinyin (with tone marks, tone numbers, or no tones). Work out which it is, then reply with one JSON object and no markdown.

Keys:
- hanzi: the word in simplified Chinese. If the input is hanzi, return it unchanged.
- pinyin: the word's standard pronunciation with tone marks, written as a dictionary writes it; neutral-tone syllables carry no mark. E.g. "lǎogōng", "lǎopo", "māma", "nǚ'ér".
- tonal_pinyin: the same pronunciation with tone numbers: lowercase, exactly one syllable per Chinese character, separated by single spaces. Write ü as ü. 儿 is always its own syllable ("er5" in erhua, "er2" when it means son or daughter). Neutral-tone syllables in the word's standard pronunciation are 5 (老婆 "lao3 po5", 爸爸 "ba4 ba5", 妈妈 "ma1 ma5"). Do not apply tone sandhi: 你好 is "ni3 hao3" and 一 is "yi1". Examples: 老公 "lao3 gong1", 女儿 "nü3 er2", 一点儿 "yi1 dian3 er5".
- english: a short English meaning.
- example_sentence_hanzi: one short, everyday sentence using the word.
- example_sentence_pinyin: that sentence's pinyin with tone marks.
- image_search: a short English photo search query for the card's picture. The picture is the learner's main clue, so it must show this word and nothing a learner could mistake for a related word that is likely to be another card. Follow the rules below.

- categories: one to three short topic categories for the word, Title Case, one or two words each, such as "Family", "Food", "Greetings", "Colors", "Verbs". Reuse names from "existing categories" whenever one fits, spelled exactly as given; add a new name only when none fits.

Rules for image_search:
- One subject, shown on its own: the thing itself, not a scene around it. For 西红柿 (tomato), "single ripe red tomato on white background", not "cutting board with chopped tomato and salad ingredients", which could be a card for cooking, salad, chopping or lettuce.
- For people and family words, show only the person the word names, with a visible cue that sets them apart from related words (husband vs wife vs couple, son vs daughter vs father or mother, grandmother vs grandfather). For 老公 (husband), "smiling groom in a suit, portrait"; for 老婆 (wife), "smiling bride in a wedding dress, portrait"; for 女儿 (daughter), "portrait of a young girl"; for 儿子 (son), "portrait of a young boy". Never pair people ("married couple cooking", "father and daughter painting").
- Leave out other people, objects and activities that are themselves likely cards: other family members, other foods, kitchens, cooking.
- For actions, one person doing only that action ("woman running on a track" for 跑步).
- For abstract words and adjectives, pick one concrete subject that clearly shows the idea and no other likely card: for 累 (tired), "exhausted man yawning, portrait"; for 高 (tall), "very tall skyscraper against blue sky".

When the input is English, choose the word native speakers use in everyday conversation: the informal, spoken word, not the formal or written one. For example "husband" is 老公 (not 丈夫), "wife" is 老婆 (not 妻子), and "dad" or "father" is 爸爸 (not 父亲). Use the formal word only when the English itself is formal or technical.

When the input is pinyin, choose the most common everyday word with that reading, matching its tones when they are given.`

// SuggestWords picks up to count everyday words for a themed set, in hanzi, avoiding the ones the
// learner already has (known; the caller filters again, since a model can miss one).
func (c *Client) SuggestWords(ctx context.Context, theme string, count int, known []string) ([]string, error) {
	user := fmt.Sprintf("theme: %s\ncount: %d", strings.TrimSpace(theme), count)
	if len(known) > 0 {
		user += "\nalready known: " + strings.Join(known, ", ")
	}
	text, err := c.complete(ctx, suggestWordsPrompt, user)
	if err != nil {
		return nil, err
	}
	var out struct {
		Words []string `json:"words"`
	}
	if err := json.Unmarshal([]byte(extractJSON(text)), &out); err != nil {
		return nil, fmt.Errorf("llm suggest words json: %w", err)
	}
	words := make([]string, 0, len(out.Words))
	for _, w := range out.Words {
		if w = strings.TrimSpace(w); w != "" {
			words = append(words, w)
		}
	}
	return words, nil
}

const suggestWordsPrompt = `You choose vocabulary for a Mandarin learner's flashcard set. Reply with one JSON object and no markdown: {"words": ["...", "..."]}.

Give exactly "count" different words or very short phrases in simplified Chinese characters that fit the theme, the most useful first. Each item is the hanzi only, no pinyin or English.
- Prefer the informal, everyday spoken word (老公 not 丈夫, 爸爸 not 父亲).
- Prefer concrete words a single photo can show clearly; avoid near-synonyms of each other in the same set.
- Never include a word listed under "already known", or the same word twice.`

// Coach turns a TonePerfect breakdown into one or two sentences of advice.
func (c *Client) Coach(ctx context.Context, hanzi string, overall float64, syllables []models.SyllableView) (string, error) {
	blob, err := json.Marshal(syllables)
	if err != nil {
		return "", err
	}
	system := `You coach a beginner learning Mandarin pronunciation. Reply in English, in one or two sentences, no preamble.
Name the character with its pinyin, e.g. 老 (lǎo), the tone it needs and what was heard, and say how to fix it. Be specific and calm.
Do not write whole sentences in Chinese. Do not mention scores, JSON, or that you are a model.`
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
