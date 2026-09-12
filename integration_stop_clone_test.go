package spx

import (
	"github.com/goplus/spx/v3/internal/coroutine"
	"testing"
)

func TestIntegratedStopAllBeforeCloneFirstSlice(t *testing.T) {
	game := setupCloneLimitGame(t)
	source := newCloneLimitSprite(game, "source")
	source.onClone = func(*cloneLimitSprite) { t.Error("clone handler ran after stop all") }
	var clone *SpriteImpl
	parent := gco.Create(&source.SpriteImpl, func(coroutine.Thread) int {
		doClone(source, nil, func(s *SpriteImpl) { clone = s })
		source.Stop(AllStop)
		return 0
	})
	gco.Join(parent)
	gco.Update()
	if !clone.isDestroyed() || game.shapeMgr.cloneCount != 0 {
		t.Fatal("pending clone survived stop all")
	}
	flushCloneProxyUpdates(game)
	if clone.runtimeState.SyncSprite != nil {
		t.Fatal("stopped clone retained its proxy")
	}
}
