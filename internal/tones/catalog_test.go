package tones

import (
	"testing"

	"github.com/csotherden/mandarin-flashcards/internal/pinyin"
)

func TestCatalog(t *testing.T) {
	lessons := map[string]bool{}
	for _, l := range Lessons {
		if lessons[l.Key] {
			t.Errorf("duplicate lesson %s", l.Key)
		}
		lessons[l.Key] = true
		if len(l.Items) < 5 || len(l.Explain) == 0 {
			t.Errorf("%s: too thin", l.Key)
		}
		items := map[string]bool{}
		for _, it := range l.Items {
			if items[it.Key] {
				t.Errorf("%s: duplicate item %s", l.Key, it.Key)
			}
			items[it.Key] = true
			if !pinyin.ValidTonal(it.Hanzi, it.Tonal) {
				t.Errorf("%s/%s: tonal %q doesn't fit %s", l.Key, it.Key, it.Tonal, it.Hanzi)
			}
			if got, _ := Find(l.Key, it.Key); got == nil {
				t.Errorf("Find(%s, %s) failed", l.Key, it.Key)
			}
		}
		for _, ex := range l.Examples {
			if !pinyin.ValidTonal(ex.Hanzi, ex.Tonal) {
				t.Errorf("%s example %s: tonal %q", l.Key, ex.Hanzi, ex.Tonal)
			}
		}
	}
	if l, it := Find("tone-1", "nope"); l == nil || it != nil {
		t.Error("unknown item should find the lesson only")
	}
}
