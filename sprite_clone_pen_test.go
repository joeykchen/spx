package spx

import (
	coreproject "github.com/goplus/spx/v3/internal/core/project"
	"reflect"
	"testing"
)

func TestClonePenStateComesFromOriginal(t *testing.T) {
	game := setupCloneLimitGame(t)
	original := newCloneLimitSprite(game, "original")
	original.pen().penWidth = 5
	original.pen().penHue = 20
	original.pen().isPenDown = true
	parent := createRuntimeClone(&original.SpriteImpl)
	parent.pen().penWidth = 50
	parent.pen().penHue = 80
	parent.pen().isPenDown = false
	original.pen().penWidth = 7
	child := createRuntimeClone(parent)
	if child.pen().penWidth != 7 || child.pen().penHue != 20 || !child.pen().isPenDown {
		t.Fatal("descendant inherited immediate parent instead of current original pen state")
	}
	if child.pen().penObj != nil {
		t.Fatal("cloning allocated or shared a drawing resource")
	}
	child.pen().penWidth = 9
	if original.pen().penWidth != 7 || parent.pen().penWidth != 50 {
		t.Fatal("clone pen state aliases an ancestor")
	}
}

func TestClonePenStateFromStageInstance(t *testing.T) {
	game := setupCloneLimitGame(t)
	template := newCloneLimitSprite(game, "template")
	template.pen().penWidth = 5
	var instance cloneLimitSprite
	applySprite(reflect.ValueOf(&instance).Elem(), template, coreproject.StageShape{})
	game.addShape(&instance.SpriteImpl)
	instance.pen().penWidth = 10
	child := createRuntimeClone(&instance.SpriteImpl)
	if child.pen().penWidth != 10 {
		t.Fatal("clone used template state instead of its original instance")
	}
}

func TestStandaloneComponentsCloneWithoutRegistry(t *testing.T) {
	sourceSprite, target := &SpriteImpl{}, &SpriteImpl{}
	sound := &soundComponent{componentBase: componentBase{sprite: sourceSprite}}
	clonedSound := sound.cloneFrom(sound, target).(*soundComponent)
	if clonedSound.sprite != target || clonedSound.soundObj != 0 {
		t.Fatal("standalone sound clone acquired resources")
	}
	pen := &penComponent{componentBase: componentBase{sprite: sourceSprite}}
	pen.penWidth = 7
	clonedPen := pen.cloneFrom(pen, target).(*penComponent)
	if clonedPen.sprite != target || clonedPen.penWidth != 7 {
		t.Fatal("standalone pen clone lost state")
	}
}
