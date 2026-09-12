package spx

import (
	"github.com/goplus/spx/v3/internal/coroutine"
	"testing"
)

func TestScratchMissingTargets(t *testing.T) {
	for _, action := range []struct {
		name string
		run  func(*cloneLimitSprite)
	}{
		{"step", func(s *cloneLimitSprite) { s.StepTo__1("missing") }},
		{"glide", func(s *cloneLimitSprite) { s.Glide__1("missing", 0) }},
		{"turn", func(s *cloneLimitSprite) { s.TurnTo__1("missing") }},
	} {
		t.Run(action.name, func(t *testing.T) {
			game := setupCloneLimitGame(t)
			sprite := newCloneLimitSprite(game, "source")
			sprite.transform().x, sprite.transform().y, sprite.transform().direction = 30, 40, 60
			before := sprite.spriteState.DirtyVersion
			finished := false
			th := gco.Create(sprite, func(coroutine.Thread) int { action.run(sprite); finished = true; return 0 })
			gco.Join(th)
			if !finished || sprite.Xpos() != 30 || sprite.Ypos() != 40 || sprite.Heading() != 60 || sprite.spriteState.DirtyVersion != before {
				t.Fatalf("missing target altered sprite: finished=%v position=(%v,%v) heading=%v dirty=%v", finished, sprite.Xpos(), sprite.Ypos(), sprite.Heading(), sprite.spriteState.DirtyVersion)
			}
		})
	}
	game := setupCloneLimitGame(t)
	sprite := newCloneLimitSprite(game, "source")
	sprite.transform().x, sprite.transform().y = 30, 40
	if got := sprite.DistanceTo__1("missing"); got != 10000 {
		t.Errorf("missing target distance=%v", got)
	}
	target := newCloneLimitSprite(game, "target")
	if got := sprite.DistanceTo__1("target"); got != 50 {
		t.Errorf("origin target distance=%v", got)
	}
	clone := createRuntimeClone(&target.SpriteImpl)
	clone.transform().x = 300
	if got := sprite.DistanceTo__1("target"); got != 50 {
		t.Errorf("distance chose a clone: %v", got)
	}
}

func TestMissingTargetDoesNotStartTimedGlide(t *testing.T) {
	game := setupCloneLimitGame(t)
	sprite := newCloneLimitSprite(game, "source")
	finished := false
	thread := gco.Create(sprite, func(coroutine.Thread) int { sprite.Glide__1("missing", 60); finished = true; return 0 })
	gco.JoinYieldedOrDone(thread)
	if !finished {
		t.Fatal("missing target started a timed glide")
	}
	var absent *cloneLimitSprite
	if sprite.DistanceTo__0(absent) != 10000 {
		t.Fatal("nil sprite was not treated as missing")
	}
}
