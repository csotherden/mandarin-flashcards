package toneperfect

import "testing"

func TestParseCompactSandhiIsNotAMiss(t *testing.T) {
	payload := []byte(`{
		"assessable": true,
		"scores": {"overall": 99.6, "tone": 98},
		"words": [
			{
				"text": "你",
				"pinyin": "ni3",
				"score": 99.3,
				"units": [
					{"dimension": "initial", "expected": "n", "heard": "n", "decision": "no_correction"},
					{"dimension": "final", "expected": "i", "heard": "i", "decision": "no_correction"},
					{"dimension": "tone", "expected": "3", "heard": "2", "decision": "no_correction"}
				]
			},
			{
				"text": "好",
				"pinyin": "hao3",
				"score": 70,
				"units": [
					{"dimension": "initial", "expected": "h", "heard": "h", "decision": "no_correction"},
					{"dimension": "final", "expected": "ao", "heard": "ang", "decision": "review"},
					{"dimension": "tone", "expected": 3, "heard": 2, "decision": "review"}
				]
			}
		]
	}`)
	got, err := ParseCompact(payload)
	if err != nil {
		t.Fatal(err)
	}
	if got.Overall != 99.6 || got.ToneAccuracy == nil || *got.ToneAccuracy != 98 {
		t.Fatalf("scores %+v", got)
	}
	if len(got.Syllables) != 2 {
		t.Fatalf("syllables %d", len(got.Syllables))
	}
	if got.Syllables[0].Status != "ok" {
		t.Fatalf("sandhi syllable status %s", got.Syllables[0].Status)
	}
	if got.Syllables[1].Status != "mismatch" {
		t.Fatalf("mismatch status %s", got.Syllables[1].Status)
	}
	if got.Syllables[1].ExpectedTone != "3" || got.Syllables[1].HeardTone != "2" {
		t.Fatalf("tones %s %s", got.Syllables[1].ExpectedTone, got.Syllables[1].HeardTone)
	}
}
