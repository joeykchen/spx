package spx

import "testing"

func TestScratchDynamicCostumeResolution(t *testing.T) {
	sprite := newTestRenderSprite()
	for _, tt := range []struct {
		value string
		want  int
	}{
		{"0", 1}, {"-1", 0}, {"-2", 1}, {"2.5", 0}, {"1.49", 0}, {" 2 ", 1},
		{"0x2", 1}, {"1e999", 0}, {"next costume", 0}, {"previous costume", 0},
		{"missing", Invalid}, {" ", Invalid}, {"--1", Invalid}, {"NaN", Invalid},
	} {
		if got := sprite.ResolveCostumeIndex(tt.value); got != tt.want {
			t.Errorf("resolve(%q)=%d, want %d", tt.value, got, tt.want)
		}
	}
	sprite.costumes[0].name = "2.5"
	if got := sprite.ResolveCostumeIndex("2.5"); got != 0 {
		t.Fatalf("name lost precedence: %d", got)
	}
}
