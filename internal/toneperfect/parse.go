package toneperfect

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/csotherden/mandarin-flashcards/internal/models"
)

type compactResponse struct {
	Assessable bool    `json:"assessable"`
	Reason     string  `json:"reason"`
	Scores     *scores `json:"scores"`
	Words      []word  `json:"words"`
}

type scores struct {
	Overall *float64 `json:"overall"`
	Tone    *float64 `json:"tone"`
}

type word struct {
	Text   string  `json:"text"`
	Word   string  `json:"word"`
	Pinyin string  `json:"pinyin"`
	Score  float64 `json:"score"`
	Units  []unit  `json:"units"`
	Phones []unit  `json:"phones"`
}

type unit struct {
	Dimension string    `json:"dimension"`
	Expected  flexValue `json:"expected"`
	Heard     flexValue `json:"heard"`
	Decision  string    `json:"decision"`
}

type flexValue string

func (f *flexValue) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*f = ""
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*f = flexValue(text)
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err == nil {
		*f = flexValue(number.String())
		return nil
	}
	return fmt.Errorf("unsupported value %s", data)
}

func syllablesFrom(doc compactResponse) []models.SyllableView {
	out := make([]models.SyllableView, 0, len(doc.Words))
	for _, item := range doc.Words {
		character := item.Text
		if character == "" {
			character = item.Word
		}
		units := item.Units
		if len(units) == 0 {
			units = item.Phones
		}
		view := models.SyllableView{
			Character: character,
			Pinyin:    item.Pinyin,
			Score:     item.Score,
			Status:    "ok",
		}
		var notes []string
		toneMismatch := false
		segmentMismatch := false
		uncertain := false
		for _, unit := range units {
			ok, unsure := unitStatus(unit.Decision, string(unit.Expected), string(unit.Heard))
			switch unit.Dimension {
			case "tone":
				view.ExpectedTone = string(unit.Expected)
				view.HeardTone = string(unit.Heard)
				if !ok {
					toneMismatch = true
					if view.ExpectedTone != "" || view.HeardTone != "" {
						notes = append(notes, fmt.Sprintf("Tone %s → heard %s", displayTone(view.ExpectedTone), displayTone(view.HeardTone)))
					}
				}
			case "initial":
				if !ok {
					segmentMismatch = true
					notes = append(notes, fmt.Sprintf("Initial %s → heard %s", unit.Expected, unit.Heard))
				}
			case "final":
				if !ok {
					segmentMismatch = true
					notes = append(notes, fmt.Sprintf("Final %s → heard %s", unit.Expected, unit.Heard))
				}
			}
			if unsure {
				uncertain = true
			}
		}
		switch {
		case segmentMismatch:
			view.Status = "mismatch"
		case toneMismatch || uncertain:
			view.Status = "warn"
		}
		view.Detail = strings.Join(notes, ". ")
		out = append(out, view)
	}
	return out
}

func unitStatus(decision, expected, heard string) (ok bool, uncertain bool) {
	switch decision {
	case "no_correction":
		return true, false
	case "uncertain":
		return false, true
	case "review":
		return false, false
	default:
		return expected == heard, false
	}
}

func displayTone(tone string) string {
	if tone == "" {
		return "?"
	}
	return tone
}
