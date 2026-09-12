package spx

import (
	corestate "github.com/goplus/spx/v3/internal/core/state"
	"github.com/goplus/spx/v3/internal/engine"
	"github.com/goplus/spx/v3/internal/ui"
)

type questionView struct {
	request *pendingQuestion
	panel   corestate.AskPanel
}

func (p *Game) ask(sprite *SpriteImpl, text string) {
	request := &pendingQuestion{
		ctx:    engine.GetCurrentThreadContext(),
		text:   text,
		sprite: sprite,
		bubble: sprite != nil && sprite.Visible(),
	}
	p.questions.add(request)
	defer p.questions.remove(request)
	for {
		done, canceled := p.questions.status(request)
		if canceled {
			gco.Abort()
		}
		if done {
			return
		}
		engine.WaitNextFrame()
	}
}

// syncQuestions runs with visual synchronization, never in a canceled script's
// cleanup. Only the front request owns the panel and its optional speech bubble.
func (p *Game) syncQuestions() {
	next := p.questions.front()
	view := &p.questionView
	if next == view.request {
		return
	}
	if previous := view.request; previous != nil {
		view.panel.Hide()
		if previous.bubble {
			previous.sprite.doStopText()
		}
	}
	*view = questionView{request: next}
	if next == nil {
		return
	}
	if p.dialogState.AskPanel == nil {
		p.dialogState.AskPanel = ui.NewUiAsk()
		p.addShape(p.dialogState.AskPanel)
	}
	view.panel = p.dialogState.AskPanel
	if next.bubble {
		next.sprite.sayOrThink(next.text, ui.StyleSay)
	}
	view.panel.Show(next.bubble, next.text, func(answer string) { p.answerQuestion(next, answer) })
}

func (p *Game) updateQuestions() {
	p.syncQuestions()
	if p.questionView.request != nil {
		p.questionView.panel.Update()
	}
}

func (p *Game) cancelQuestions() {
	p.questions.clear(false)
}

// Retire question speech before publishing completion: the resumed script can
// immediately display new speech, which later panel cleanup must leave intact.
func (p *Game) answerQuestion(request *pendingQuestion, answer string) {
	if !p.questions.isCurrent(request) {
		return
	}
	if request.bubble {
		request.bubble = false
		request.sprite.doStopText()
	}
	p.questions.submit(request, answer)
}
