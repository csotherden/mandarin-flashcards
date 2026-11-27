// Package pinyin checks the numbered pinyin cards carry for TonePerfect.
package pinyin

import (
	"regexp"
	"strings"
	"unicode"
)

// numbered is one numbered syllable: letters (ü written as ü, v or u:) and a tone 1 to 5, 5 being
// the neutral tone.
var numbered = regexp.MustCompile(`^[a-zü:]+[1-5]$`)

// HanCount is the number of Han characters in s (punctuation and Latin letters are not read).
func HanCount(s string) int {
	n := 0
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			n++
		}
	}
	return n
}

// ValidTonal reports whether tonal is numbered pinyin TonePerfect accepts for hanzi: one
// space-separated syllable per Han character, each with a tone number ("lao3 po5", "yi1 dian3 er5").
func ValidTonal(hanzi, tonal string) bool {
	syllables := strings.Fields(strings.ToLower(tonal))
	if len(syllables) == 0 || len(syllables) != HanCount(hanzi) {
		return false
	}
	for _, s := range syllables {
		if !numbered.MatchString(s) {
			return false
		}
	}
	return true
}

// NormalizeTonal lowercases tonal and collapses its spacing.
func NormalizeTonal(tonal string) string {
	return strings.Join(strings.Fields(strings.ToLower(tonal)), " ")
}
