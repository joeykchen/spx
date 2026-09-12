package spx

import "testing"

func TestScratchListContents(t *testing.T) {
	for _, tt := range []struct {
		name   string
		values []any
		want   string
	}{
		{"digits", []any{1, 2}, "1 2"}, {"strings", []any{"1", "2"}, "12"},
		{"mixed", []any{"1", 2}, "1 2"}, {"unicode", []any{"你", "好"}, "你好"},
		{"astral", []any{"😀", "a"}, "😀 a"}, {"empty item", []any{"", "a"}, " a"},
		{"combining", []any{"e\u0301", "a"}, "e\u0301 a"}, {"bool", []any{true, false}, "true false"},
		{"wrapped", []any{NewValue("你"), NewValue("好")}, "你好"}, {"empty", nil, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			list := NewList(tt.values...)
			if got := list.String(); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
