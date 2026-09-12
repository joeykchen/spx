package spx

import (
	"context"
	"github.com/goplus/spx/v3/internal/coroutine"
	itime "github.com/goplus/spx/v3/internal/time"
	"slices"
	"sync"
	"testing"
)

type questionPanelSpy struct {
	shown    []string
	inBubble []bool
	callback func(string)
	visible  bool
}

func (p *questionPanelSpy) Show(inBubble bool, text string, callback func(string)) {
	p.shown = append(p.shown, text)
	p.inBubble = append(p.inBubble, inBubble)
	p.callback = callback
	p.visible = true
}
func (p *questionPanelSpy) Update() {}
func (p *questionPanelSpy) Hide()   { p.visible = false; p.callback = nil }

func TestScratchQuestionsAreFIFO(t *testing.T) {
	co, game := setupRuntimeEventGame(t)
	panel := &questionPanelSpy{}
	game.dialogState.AskPanel = panel
	var finished []string
	first := co.Create(game, func(coroutine.Thread) int { game.Ask("first"); finished = append(finished, "first"); return 0 })
	co.JoinYieldedOrDone(first)
	game.syncQuestions()
	second := co.Create(game, func(coroutine.Thread) int { game.Ask("second"); finished = append(finished, "second"); return 0 })
	co.JoinYieldedOrDone(second)
	game.syncQuestions()
	if !slices.Equal(panel.shown, []string{"first"}) {
		t.Fatalf("questions overwrote active prompt: %v", panel.shown)
	}
	oldCallback := panel.callback
	panel.callback("one")
	itime.Update(0, 0)
	co.Update()
	game.syncQuestions()
	if !slices.Equal(finished, []string{"first"}) {
		t.Fatalf("answered wrong question: %v", finished)
	}
	if !slices.Equal(panel.shown, []string{"first", "second"}) {
		t.Fatalf("queue did not advance: %v", panel.shown)
	}
	oldCallback("stale")
	if game.Answer() != "one" {
		t.Fatal("old callback overwrote latest answer")
	}
	panel.callback("two")
	itime.Update(0, 0)
	co.Update()
	game.syncQuestions()
	if !slices.Equal(finished, []string{"first", "second"}) || game.Answer() != "two" || panel.visible {
		t.Fatal("queue did not finish cleanly")
	}

}

func TestScratchQuestionCancellation(t *testing.T) {
	for _, cancelFirst := range []bool{false, true} {
		t.Run(map[bool]string{false: "queued", true: "active"}[cancelFirst], func(t *testing.T) {
			co, game := setupRuntimeEventGame(t)
			panel := &questionPanelSpy{}
			game.dialogState.AskPanel = panel
			first := co.Create(game, func(coroutine.Thread) int { game.Ask("first"); return 0 })
			co.JoinYieldedOrDone(first)
			second := co.Create(game, func(coroutine.Thread) int { game.Ask("second"); return 0 })
			co.JoinYieldedOrDone(second)
			game.syncQuestions()
			stale := panel.callback
			canceled := second
			if cancelFirst {
				canceled = first
			}
			co.StopIf(func(th coroutine.Thread) bool { return th == canceled })
			if cancelFirst {
				stale("canceled before visual sync")
				if game.Answer() != "" {
					t.Fatal("canceled request accepted an answer before UI cleanup")
				}
			}
			co.Update()
			game.syncQuestions()
			expected := []string{"first"}
			if cancelFirst {
				expected = append(expected, "second")
			}
			if !slices.Equal(panel.shown, expected) {
				t.Fatalf("cancel changed prompt order: %v", panel.shown)
			}
			if cancelFirst {
				stale("stale")
				if game.Answer() != "" {
					t.Fatal("canceled callback changed answer")
				}
			}
			game.stopAllResources()
			game.syncQuestions()
			if panel.visible {
				t.Fatal("stop all left question visible")
			}
			stale("late")
			if game.Answer() != "" {
				t.Fatal("stop all retained an active answer callback")
			}
		})
	}
}

func TestHiddenSpriteQuestionUsesPanel(t *testing.T) {
	game := setupCloneLimitGame(t)
	panel := &questionPanelSpy{}
	game.dialogState.AskPanel = panel
	sprite := newCloneLimitSprite(game, "hidden")
	sprite.spriteState.IsVisible = false
	th := gco.Create(sprite, func(coroutine.Thread) int { sprite.Ask("hidden question"); return 0 })
	gco.JoinYieldedOrDone(th)
	game.syncQuestions()
	if !slices.Equal(panel.inBubble, []bool{false}) || sprite.components.bubble != nil {
		t.Fatal("hidden sprite used an invisible speech bubble")
	}
}

func TestQuestionQueueAnswerLifetime(t *testing.T) {
	var q questionQueue
	first := &pendingQuestion{ctx: context.Background()}
	q.add(first)
	q.submit(first, "kept")
	second := &pendingQuestion{ctx: context.Background()}
	q.add(second)
	q.clear(false)
	if q.answerText() != "kept" {
		t.Fatal("stop all cleared the prior answer")
	}
	if done, canceled := q.status(second); done || !canceled {
		t.Fatal("stopped question completed as an answer")
	}
	q.clear(true)
	if q.answerText() != "" {
		t.Fatal("project reset retained the prior answer")
	}
}

func TestQuestionQueueConcurrentAnswerAndCancellation(t *testing.T) {
	for range 32 {
		var q questionQueue
		ctx, cancel := context.WithCancel(context.Background())
		request := &pendingQuestion{ctx: ctx}
		q.add(request)
		var workers sync.WaitGroup
		workers.Go(func() { q.submit(request, "answer") })
		workers.Go(func() { cancel(); q.front() })
		workers.Go(func() { q.answerText(); q.status(request) })
		workers.Wait()
		answer := q.answerText()
		q.submit(request, "stale")
		if q.front() != nil || q.answerText() != answer {
			t.Fatal("canceled request remained answerable")
		}
	}
}

func TestQuestionAnswerPreservesFollowingSpeech(t *testing.T) {
	co, game := setupRuntimeEventGame(t)
	panel := &questionPanelSpy{}
	game.dialogState.AskPanel = panel
	sprite := &SpriteImpl{g: game}
	sprite.components.bubble = &bubbleComponent{componentBase: componentBase{sprite: sprite}, textObj: &textBubble{}}
	followingSpeech := &textBubble{msg: "answer received"}
	thread := co.Create(game, func(coroutine.Thread) int {
		game.Ask("question")
		if sprite.components.bubble.textObj != nil {
			t.Error("question bubble remained when script resumed")
		}
		sprite.components.bubble.textObj = followingSpeech
		return 0
	})
	co.JoinYieldedOrDone(thread)
	game.syncQuestions()
	// Attach a displayed bubble without creating a native UI scene.
	request := game.questions.front()
	request.sprite, request.bubble = sprite, true
	panel.callback("answer")
	itime.Update(0, 0)
	co.Update()
	game.syncQuestions()
	if sprite.components.bubble.textObj != followingSpeech {
		t.Fatal("question cleanup removed the script's following speech")
	}
}
