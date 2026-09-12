package ui

import (
	"github.com/goplus/spx/v3/internal/enginewrap"
	gdx "github.com/goplus/spx/v3/pkg/spx/pkg/engine"
	"testing"
)

type askInputSpy struct {
	gdx.IInputMgr
	pressed bool
}

func (p *askInputSpy) GetKey(key int64) bool { return p.pressed && key == int64(gdx.KeyEnter) }

type askRenderSpy struct {
	gdx.IUiMgr
	visible bool
}

func (p *askRenderSpy) SetText(gdx.Object, string)      {}
func (p *askRenderSpy) GetText(gdx.Object) string       { return "answer" }
func (p *askRenderSpy) SetVisible(_ gdx.Object, v bool) { p.visible = v }

func TestUiAskConsumesAnswerOnceAndRequiresNewEnter(t *testing.T) {
	input, render := &askInputSpy{}, &askRenderSpy{}
	oldInput, oldUI := gdx.InputMgr, gdx.UiMgr
	gdx.InputMgr, gdx.UiMgr = input, render
	enginewrap.Init(func(f func()) { f() })
	t.Cleanup(func() { gdx.InputMgr, gdx.UiMgr = oldInput, oldUI; enginewrap.Init(nil) })
	panel := &UiAsk{input: &UiNode{}, askBody: &UiNode{}, askLabel: &UiNode{}}
	answered := 0
	panel.Show(false, "first", func(string) { answered++ })
	input.pressed = true
	panel.Update()
	if answered != 1 {
		t.Fatalf("first answer count=%d", answered)
	}
	panel.Show(false, "second", func(string) { answered++ })
	panel.Update()
	if answered != 1 {
		t.Fatalf("held Enter answered next question: %d", answered)
	}
	input.pressed = false
	panel.Update()
	input.pressed = true
	panel.Update()
	if answered != 2 {
		t.Fatalf("fresh Enter did not answer: %d", answered)
	}
	panel.handleCheck()
	panel.handleCheck()
	if answered != 2 {
		t.Fatalf("duplicate callbacks submitted an answer: %d", answered)
	}
	panel.Show(false, "canceled", func(string) { answered++ })
	panel.Hide()
	panel.handleCheck()
	if panel.OnCheck != nil || answered != 2 {
		t.Fatal("hidden dialog retained its callback")
	}

}
