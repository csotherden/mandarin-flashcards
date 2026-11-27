package generate

import (
	"reflect"
	"testing"
)

func TestPickNew(t *testing.T) {
	chosen, skipped := pickNew([]string{"老公", "妈妈", "bad word", "妈妈", "爸爸", "奶奶", "爷爷"}, []string{"老公"}, 3)
	if !reflect.DeepEqual(chosen, []string{"妈妈", "爸爸", "奶奶"}) {
		t.Errorf("chosen = %v", chosen)
	}
	if !reflect.DeepEqual(skipped, []string{"老公", "妈妈"}) {
		t.Errorf("skipped = %v (a known word and a repeat)", skipped)
	}
}

func TestTitleCase(t *testing.T) {
	if got := titleCase("ordering  coffee"); got != "Ordering Coffee" {
		t.Errorf("titleCase = %q", got)
	}
}
