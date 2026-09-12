package spx

import (
	"reflect"
	"testing"

	"github.com/goplus/spx/v3/internal/coroutine"
)

func TestCloneCreationPreservesParentSlice(t *testing.T) {
	game := setupCloneLimitGame(t)
	source := newCloneLimitSprite(game, "source")
	shared := 0
	var order []int
	source.onClone = func(*cloneLimitSprite) { order = append(order, shared) }
	var clone *SpriteImpl
	parent := gco.Create(&source.SpriteImpl, func(coroutine.Thread) int {
		doClone(source, nil, func(s *SpriteImpl) { clone = s })
		shared = 1
		if clone.isCloneProxyPublicationReady() {
			t.Error("clone published before its handler ran")
		}
		return 0
	})
	gco.Join(parent)
	gco.Update()
	if !reflect.DeepEqual(order, []int{1}) {
		t.Fatalf("clone saw %v, want parent's updated value [1]", order)
	}
	if !clone.isCloneProxyPublicationReady() {
		t.Fatal("clone was not ready after its first slice")
	}
}

func TestClonePublicationWaitsForEveryHandlerFirstSlice(t *testing.T) {
	game := setupCloneLimitGame(t)
	source := newCloneLimitSprite(game, "source")
	gate := gco.NewLatch()
	var order []int
	source.onInit = func(s *cloneLimitSprite) {
		s.OnCloned__0(func() { order = append(order, 1); gate.Wait() })
	}
	source.onClone = func(*cloneLimitSprite) { order = append(order, 2); gate.Wait() }
	var clone *SpriteImpl
	parent := gco.Create(&source.SpriteImpl, func(coroutine.Thread) int {
		doClone(source, nil, func(s *SpriteImpl) { clone = s })
		if len(order) != 0 {
			t.Error("clone handlers interrupted parent slice")
		}
		flushCloneProxyUpdates(game)
		if clone.cloneProxyPublicationState() != cloneProxyPending {
			t.Error("render published an uninitialized clone")
		}
		return 0
	})
	gco.Join(parent)
	gco.Update()
	if !reflect.DeepEqual(order, []int{1, 2}) {
		t.Fatalf("handler order = %v", order)
	}
	if !clone.isCloneProxyPublicationReady() {
		t.Fatal("publication waited for entire handler bodies")
	}
	flushCloneProxyUpdates(game)
	if clone.isCloneProxyPublicationBlocked() {
		t.Fatal("initialized clone did not publish")
	}
	gate.Open()
	gco.Update()
}

func TestCloneCanceledBeforeFirstSliceDoesNotStrandPublication(t *testing.T) {
	game := setupCloneLimitGame(t)
	source := newCloneLimitSprite(game, "source")
	source.onClone = func(*cloneLimitSprite) { t.Error("canceled handler ran") }
	var clone *SpriteImpl
	parent := gco.Create(&source.SpriteImpl, func(coroutine.Thread) int {
		doClone(source, nil, func(s *SpriteImpl) { clone = s })
		clone.Stop(ThisSprite)
		return 0
	})
	gco.Join(parent)
	updateRuntimeEventSchedulerUntil(t, gco, clone.isCloneProxyPublicationReady)
}

func TestCloneRejectedHandlersDoNotStrandPublication(t *testing.T) {
	game := setupCloneLimitGame(t)
	source := newCloneLimitSprite(game, "source")
	source.onClone = func(*cloneLimitSprite) { t.Error("rejected handler ran") }
	var clone *SpriteImpl
	parent := gco.Create(&source.SpriteImpl, func(thread coroutine.Thread) int {
		gco.Stop(thread)
		doClone(source, nil, func(s *SpriteImpl) { clone = s })
		return 0
	})
	gco.Join(parent)
	gco.Update()
	if clone == nil || !clone.isCloneProxyPublicationReady() {
		t.Fatal("rejected batch stranded clone publication")
	}
}
