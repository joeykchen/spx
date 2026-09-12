package spx

import (
	"testing"

	coreevent "github.com/goplus/spx/v3/internal/core/event"
	"github.com/goplus/spx/v3/internal/coroutine"
	"github.com/goplus/spx/v3/internal/engine"
)

type stopAllAudioBackend struct {
	animationAudioBackend
	stopped   int
	destroyed []engine.Object
}

func (b *stopAllAudioBackend) StopAll()                       { b.stopped++; clear(b.playing) }
func (b *stopAllAudioBackend) DestroyAudio(obj engine.Object) { b.destroyed = append(b.destroyed, obj) }

func TestStopAllCleansResourcesBeforeAbortingCaller(t *testing.T) {
	for _, fromClone := range []bool{false, true} {
		t.Run(map[bool]string{false: "stage", true: "clone"}[fromClone], func(t *testing.T) {
			game := setupCloneLimitGame(t)
			game.scriptEventBindings.init(&game.scriptEvents, game)
			backend := &stopAllAudioBackend{}
			game.soundMgr.Init(backend)
			source := newCloneLimitSprite(game, "source")
			source.transform().x = 42
			source.onInit = func(s *cloneLimitSprite) { s.OnKey__0(KeySpace, func() {}) }
			var clone *SpriteImpl
			doClone(source, nil, func(s *SpriteImpl) { clone = s })
			clone.sound().soundObj = 17
			playback := game.soundMgr.Play(17, "long.wav", false, false, 0, 0, 0)
			gate := gco.NewLatch()
			peerFinished := false
			peer := gco.Create(&source.SpriteImpl, func(coroutine.Thread) int { gate.Wait(); peerFinished = true; return 0 })
			gco.JoinYieldedOrDone(peer)
			caller := threadObj(game)
			if fromClone {
				caller = clone
			}
			afterStop := false
			stop := gco.Create(caller, func(coroutine.Thread) int {
				if fromClone {
					clone.Stop(AllStop)
				} else {
					game.Stop(AllStop)
				}
				afterStop = true
				return 0
			})
			gco.Join(stop)
			gco.Join(peer)
			if afterStop || peerFinished {
				t.Fatal("stopped script continued")
			}
			if !clone.isDestroyed() || game.shapeMgr.cloneCount != 0 || game.shapeMgr.findShapeIndex(clone) >= 0 {
				t.Fatal("stop all retained clone or its quota")
			}
			if len(game.scriptEvents.manager.Snapshot(coreevent.BucketKeyPressed)) != 0 {
				t.Fatal("clone event registration survived")
			}
			if backend.IsPlaying(playback) || backend.stopped != 1 {
				t.Fatal("stop all left audio playing")
			}
			if len(backend.destroyed) != 1 || backend.destroyed[0] != 17 {
				t.Fatalf("released sound objects = %v", backend.destroyed)
			}
			if source.isDestroyed() || source.transform().x != 42 || len(game.getAllShapes()) != 1 {
				t.Fatal("stop all reset or removed original")
			}
			// Repeating the stop must not destroy a clone or release its slot twice.
			again := gco.Create(game, func(coroutine.Thread) int { game.Stop(AllStop); return 0 })
			gco.Join(again)
			if game.shapeMgr.cloneCount != 0 || len(backend.destroyed) != 1 {
				t.Fatal("repeated stop duplicated cleanup")
			}
		})
	}
}

func TestStopAllCancelsWaitingCloneHandler(t *testing.T) {
	game := setupCloneLimitGame(t)
	source := newCloneLimitSprite(game, "source")
	gate := gco.NewLatch()
	source.onClone = func(*cloneLimitSprite) { gate.Wait(); t.Error("clone handler continued after stop all") }
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

func TestStopAllClearsExistingSoundEffectsWithoutAllocating(t *testing.T) {
	co, game := setupRuntimeEventGame(t)
	backend := &fakeAudioBackend{pan: 0.5, pitch: 2}
	game.soundMgr.Init(backend)
	game.audioState.SoundObj = 9
	stop := co.Create(game, func(coroutine.Thread) int { game.Stop(AllStop); return 0 })
	co.Join(stop)
	if backend.pan != 0 || backend.pitch != 1 || backend.createCalls != 0 {
		t.Fatalf("audio after stop: %+v", backend)
	}
}
