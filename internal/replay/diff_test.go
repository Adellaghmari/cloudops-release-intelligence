package replay

import "testing"

func TestCompareKinds(t *testing.T) {
	if CompareStrings("x", "", "b").Kind != Added {
		t.Fatal("added")
	}
	if CompareStrings("x", "a", "").Kind != Removed {
		t.Fatal("removed")
	}
	if CompareStrings("x", "a", "b").Kind != Changed {
		t.Fatal("changed")
	}
	if CompareStrings("x", "a", "a").Kind != Unchanged {
		t.Fatal("unchanged")
	}
	if CompareInts("n", 1, 1).Kind != Unchanged || CompareBools("f", true, false).Kind != Changed {
		t.Fatal("typed")
	}
}
