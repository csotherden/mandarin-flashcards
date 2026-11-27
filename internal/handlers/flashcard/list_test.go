package flashcard

import (
	"strings"
	"testing"

	"github.com/csotherden/mandarin-flashcards/internal/models"
)

func f(v float64) *float64 { return &v }

func hanzi(cards []models.FlashcardResponse) string {
	out := make([]string, len(cards))
	for i, c := range cards {
		out[i] = c.Hanzi
	}
	return strings.Join(out, " ")
}

func TestSortCards(t *testing.T) {
	base := func() []models.FlashcardResponse {
		return []models.FlashcardResponse{
			{Hanzi: "香蕉", TonalPinyin: "xiang1 jiao1", RecentScore: f(80)},
			{Hanzi: "苹果", TonalPinyin: "ping2 guo3"},
			{Hanzi: "爸爸", TonalPinyin: "ba4 ba5", RecentScore: f(40)},
			{Hanzi: "女儿", TonalPinyin: "nü3 er2", RecentScore: f(95)},
		}
	}
	cards := base()
	SortCards(cards, "az")
	if got := hanzi(cards); got != "爸爸 女儿 苹果 香蕉" {
		t.Errorf("az = %s", got)
	}
	cards = base()
	SortCards(cards, "weakest")
	if got := hanzi(cards); got != "爸爸 香蕉 女儿 苹果" {
		t.Errorf("weakest = %s (unpracticed last)", got)
	}
	cards = base()
	SortCards(cards, "newest")
	if got := hanzi(cards); got != "香蕉 苹果 爸爸 女儿" {
		t.Errorf("newest should keep query order, got %s", got)
	}
}
