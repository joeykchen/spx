package spx

import (
	"fmt"
	"github.com/goplus/spx/v3/internal/coroutine"
	"math"
	"testing"
)

func TestScratchImmediateGlide(t *testing.T) {
	for _, duration := range []float64{0, -1, math.NaN()} {
		t.Run(fmt.Sprint(duration), func(t *testing.T) {
			game := setupCloneLimitGame(t)
			sprite := newCloneLimitSprite(game, "source")
			finished := false
			thread := gco.Create(sprite, func(coroutine.Thread) int { sprite.GlideToXYpos(30, 40, duration); finished = true; return 0 })
			gco.JoinYieldedOrDone(thread)
			if !finished || sprite.Xpos() != 30 || sprite.Ypos() != 40 {
				t.Fatalf("immediate glide: finished=%v position=(%v,%v)", finished, sprite.Xpos(), sprite.Ypos())
			}
			if sprite.animation().curTweenState != nil {
				t.Fatal("immediate glide allocated tween state")
			}
		})
	}
}
