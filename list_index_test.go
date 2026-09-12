package spx

import (
	"fmt"
	"math"
	"testing"
)

func TestScratchListIndexSeparatesNumbersFromSpecialPositions(t *testing.T) {
	for _, value := range []any{-1, -2, -3, -4, 0, 0.9, "-2", "--1", "1_0", "0x1p0", "+0x1", "1e", "\u00852\u0085", "ALL", " last ", "", nil, false, math.NaN(), math.Inf(1), math.Inf(-1), 1e30} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			index := ScratchListIndex(value)
			if index != Invalid {
				t.Fatalf("ScratchListIndex(%v) = %d, want Invalid", value, index)
			}
			list := NewList("a", "b")
			list.Delete(index)
			list.Insert(index, "x")
			list.Set(index, "x")
			if list.String() != "ab" || list.At(index).String() != "" {
				t.Fatal("invalid index affected list")
			}
		})
	}
	for _, tt := range []struct {
		value any
		want  Pos
	}{
		{1, 0}, {2.9, 1}, {"2.9", 1}, {"  +2e0 ", 1}, {"0x10", 15}, {"0o10", 7}, {"0b10", 1},
		{"\ufeff2\ufeff", 1}, {true, 0}, {NewValue("2"), 1}, {"all", All}, {"last", Last}, {"random", Random}, {"any", Random},
	} {
		if got := ScratchListIndex(tt.value); got != tt.want {
			t.Errorf("ScratchListIndex(%v) = %d, want %d", tt.value, got, tt.want)
		}
	}
}

func TestScratchListIndexRespectsOperationBounds(t *testing.T) {
	list := NewList("a", "b")
	list.Insert(ScratchListIndex(999), "x")
	list.Insert(ScratchListIndex("all"), "x")
	if list.String() != "ab" {
		t.Fatal("invalid insertion changed the list")
	}
	list.Insert(ScratchListIndex("last"), "c")
	if list.String() != "abc" || list.At(ScratchListIndex("last")).String() != "c" {
		t.Fatal("last did not use operation's bounds")
	}
	list.Delete(ScratchListIndex("all"))
	if list.Len() != 0 {
		t.Fatal("explicit all did not clear list")
	}
}
