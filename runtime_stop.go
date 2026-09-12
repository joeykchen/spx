package spx

import (
	coreevent "github.com/goplus/spx/v3/internal/core/event"
	"github.com/goplus/spx/v3/internal/coroutine"
	"slices"
)

// Stop stops scripts selected by kind. The explicit receiver enables generated
// direct-call adapters for interpreted scripts.
func (p *Game) Stop(kind StopKind) {
	p.scriptEventBindings.Stop(kind)
}

// Stop stops scripts selected by kind.
func (p *SpriteImpl) Stop(kind StopKind) {
	p.scriptEventBindings.Stop(kind)
}

func (p *scriptEventBindings) Stop(kind StopKind) {
	// Read the receiver before the fast path to preserve nil-receiver behavior.
	owner := p.pthis
	if kind == ThisScript {
		// This signal is scoped by Procedure and never filters other threads.
		gco.AbortThisScript()
		return
	}
	if kind == AllStop {
		p.scriptEventRegistry.stopAllEpoch.Add(1)
	}

	var current coroutine.Thread
	if gco.IsInCoroutine() {
		current = gco.Current()
	}
	filter, abort := coreevent.ResolveStop(
		kind,
		owner,
		func(obj any) bool { return isSprite(obj) },
		func(obj any) bool { return isGame(obj) },
	)
	if filter != nil {
		gco.StopIf(func(th coroutine.Thread) bool {
			if !filter(th.Obj, th == current) {
				return false
			}
			return kind != AllStop || th != current && !p.scriptEventRegistry.isPendingStartThread(th)
		})
	}
	if kind == AllStop {
		if game := activeGame(); game != nil {
			game.stopAllResources()
		}
	}
	if abort {
		gco.Abort()
	}
}

// stopAllResources preserves original sprite state while disposing of clones
// and resetting the effects and playback associated with project execution.
// Other scripts have been canceled; the caller remains alive for engine calls.
func (p *Game) stopAllResources() {
	p.resetGraphicEffectsOnStopAll()
	p.StopAllSounds()
	p.clearSoundEffects(p.audioState.SoundObj)
	for _, shape := range slices.Clone(p.getAllShapes()) {
		sprite, ok := shape.(*SpriteImpl)
		if !ok {
			continue
		}
		if sprite.IsCloned() {
			sprite.destroy()
		} else if sound := sprite.components.sound; sound != nil {
			p.clearSoundEffects(sound.soundObj)
		}
	}
}
