package spx

import (
	"github.com/goplus/spbase/mathf"
	coreproject "github.com/goplus/spx/v3/internal/core/project"
	spxlog "github.com/goplus/spx/v3/internal/log"
	"math"
)

func (t *transformComponent) glide(x, y float64, secs float64) {
	// Scratch casts NaN to zero and completes non-positive glides immediately.
	if secs <= 0 || math.IsNaN(secs) {
		t.moveTo(x, y)
		return
	}

	if isDebugInstrEnabled() {
		spxlog.Debug("Glide: sprite=%s, x=%v, y=%v, secs=%v", t.sprite.name, x, y, secs)
	}

	from := mathf.NewVec2(t.x, t.y)
	to := mathf.NewVec2(x, y)

	animName := t.sprite.getStateAnimName(StateGlide)
	t.sprite.doTween(animName, &coreproject.AniConfig{
		Duration: secs,
		From:     &from,
		To:       &to,
		AniType:  coreproject.AniTypeGlide,
		IsLoop:   true,
	})
}
