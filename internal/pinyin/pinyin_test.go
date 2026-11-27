package pinyin

import "testing"

func TestValidTonal(t *testing.T) {
	cases := []struct {
		hanzi, tonal string
		want         bool
	}{
		{"妈妈", "ma1 ma5", true},
		{"一点儿", "yi1 dian3 er5", true},
		{"女儿", "nü3 er2", true},
		{"女儿", "nv3 er2", true},
		{"你好！", "ni3 hao3", true}, // punctuation is not read
		{"妈妈", "ma1ma5", false},   // run together
		{"妈妈", "ma1 ma", false},   // neutral tone without its 5
		{"妈妈", "mā ma", false},    // tone marks belong in pinyin, not tonal_pinyin
		{"一点儿", "yi1 dian3", false},
		{"妈妈", "", false},
	}
	for _, c := range cases {
		if got := ValidTonal(c.hanzi, c.tonal); got != c.want {
			t.Errorf("ValidTonal(%q, %q) = %v, want %v", c.hanzi, c.tonal, got, c.want)
		}
	}
	if got := NormalizeTonal("  Ma1   MA5 "); got != "ma1 ma5" {
		t.Errorf("NormalizeTonal = %q", got)
	}
}
