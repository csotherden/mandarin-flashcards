package autofill

import "testing"

func TestIsHanzi(t *testing.T) {
	cases := map[string]bool{
		"老公":        true,
		"你好！":       true,
		"一 点 儿":     true,
		"husband":   false,
		"lǎogōng":   false,
		"lao3gong1": false,
		"老gong":     false,
		"":          false,
	}
	for in, want := range cases {
		if got := isHanzi(in); got != want {
			t.Errorf("isHanzi(%q) = %v, want %v", in, got, want)
		}
	}
}
