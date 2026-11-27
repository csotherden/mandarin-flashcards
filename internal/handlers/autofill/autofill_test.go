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

func TestMatchCategories(t *testing.T) {
	got := matchCategories([]string{"family", "People", "Family", "", "Relationships", "Extra"}, []string{"Family"})
	want := []string{"Family", "People", "Relationships"}
	if len(got) != len(want) {
		t.Fatalf("matchCategories = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("matchCategories = %v, want %v", got, want)
		}
	}
}
