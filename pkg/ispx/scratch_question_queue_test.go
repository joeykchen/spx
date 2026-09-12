package ispx

import (
	"github.com/goplus/ixgo"
	"testing"
)

func TestInterpretedScratchQuestionQueue(t *testing.T) {
	const source = `package main
import "github.com/goplus/spx/v3"
func Run() bool {
game := &spx.Game{}
return game.Answer() == ""
}`
	ctx := ixgo.NewContext(ixgo.EnableCachedReg | ixgo.SupportMultipleInterp)
	interp, err := ctx.LoadInterp("main.go", source)
	if err != nil {
		t.Fatal(err)
	}
	defer interp.UnsafeRelease()
	got, err := interp.RunFunc("Run")
	if err != nil || got != true {
		t.Fatalf("interpreted result=%v, %v", got, err)
	}
}
