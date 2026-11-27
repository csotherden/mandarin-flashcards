package toneperfect

import (
	"strings"
	"unicode"
)

// requestPinyin is the card's pinyin when TonePerfect will accept it, else "".
//
// TonePerfect only uses pinyin to choose between a character's readings, and it refuses pinyin
// that is not one toned syllable per character (422 invalid_pinyin): run-together syllables
// ("lǎogōng") and unmarked neutral tones ("māma", "mā ma") are both refused, while the text alone
// is always fine. So the pinyin is sent only when it has exactly one syllable per Han character
// and every syllable carries a tone, as a tone mark (nǐ) or a number (ni3, ma5).
func requestPinyin(hanzi, pinyin string) string {
	pinyin = strings.TrimSpace(pinyin)
	if pinyin == "" {
		return ""
	}
	chars := 0
	for _, r := range hanzi {
		if unicode.Is(unicode.Han, r) {
			chars++
		}
	}
	syllables := strings.FieldsFunc(pinyin, func(r rune) bool {
		return unicode.IsSpace(r) || r == '\'' || r == '’'
	})
	if chars == 0 || len(syllables) != chars {
		return ""
	}
	for _, s := range syllables {
		if !hasTone(s) {
			return ""
		}
	}
	return pinyin
}

// hasTone reports whether one pinyin syllable carries a tone: a tone-marked vowel, or a trailing
// tone number 1 to 5.
func hasTone(syllable string) bool {
	if last := syllable[len(syllable)-1]; last >= '1' && last <= '5' {
		return true
	}
	return strings.ContainsAny(syllable, "āáǎàēéěèīíǐìōóǒòūúǔùǖǘǚǜĀÁǍÀĒÉĚÈĪÍǏÌŌÓǑÒŪÚǓÙǕǗǙǛ")
}
