package toneperfect

import (
	"strings"
	"testing"
)

// The cards' pinyin as people type it; TonePerfect's /v1/lint refuses the run-together and
// unmarked-neutral-tone forms with invalid_pinyin, and accepts the text alone.
func TestRequestPinyin(t *testing.T) {
	cases := []struct{ hanzi, pinyin, want string }{
		{"妈妈", "māma", ""},             // run together, neutral tone unmarked
		{"妈妈", "mā ma", ""},            // neutral tone unmarked
		{"妈妈", "ma1 ma5", "ma1 ma5"},   // numbered, neutral as 5
		{"老公", "lǎogōng", ""},          // run together
		{"老公", "lǎo gōng", "lǎo gōng"}, // one toned syllable per character
		{"女儿", "nǚ'ér", "nǚ'ér"},       // apostrophe separates syllables
		{"你好！", "nǐ hǎo", "nǐ hǎo"},    // punctuation is not a character to read
		{"你好", "nǐ", ""},               // too few syllables
		{"你好", "", ""},
	}
	for _, c := range cases {
		if got := requestPinyin(c.hanzi, c.pinyin); got != c.want {
			t.Errorf("requestPinyin(%q, %q) = %q, want %q", c.hanzi, c.pinyin, got, c.want)
		}
	}
}

func TestStatusErrorNamesTheCode(t *testing.T) {
	err := statusError(422, []byte(`{"error":{"code":"invalid_pinyin","message":"Pinyin must be syllables with tones."},"request_id":"req_1"}`))
	if !strings.Contains(err.Error(), "422 invalid_pinyin: Pinyin must be syllables with tones. (request req_1)") {
		t.Fatal(err)
	}
	if err := statusError(502, []byte("bad gateway")); err.Error() != "toneperfect status 502" {
		t.Fatal(err)
	}
}
