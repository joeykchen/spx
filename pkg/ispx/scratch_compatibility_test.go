package ispx

import (
	"testing"

	"github.com/goplus/ixgo"
)

func TestInterpretedScratchCompatibility(t *testing.T) {
	const source = `package main
import "github.com/goplus/spx/v3"
func Run() bool {
 numbers := spx.NewList(1, 2)
 letters := spx.NewList("你", "好")
 return spx.Equal("0x10", 16) && !spx.Equal("--1", 1) &&
  spx.Iround(-1.5) == -1 && numbers.String() == "1 2" && letters.String() == "你好"
}`
	ctx := ixgo.NewContext(ixgo.EnableCachedReg | ixgo.SupportMultipleInterp)
	interp, err := ctx.LoadInterp("main.go", source)
	if err != nil {
		t.Fatal(err)
	}
	defer interp.UnsafeRelease()
	got, err := interp.RunFunc("Run")
	if err != nil || got != true {
		t.Fatalf("interpreted compatibility=%v, %v", got, err)
	}
}

func TestInterpretedScratchTextAndAnswer(t *testing.T) {
	const source = `package main
import "github.com/goplus/spx/v3"
func Run() bool {
 var number spx.Value
 number.Set(1e6)
 list := spx.NewList(1e-7, "x")
 game := &spx.Game{}
 return number.String() == "1000000" && list.String() == "1e-7 x" &&
  spx.Equal("İ", "i\u0307") && spx.Contains("ΟΣ", "ος") &&
  spx.Compare("😀", "\ue000") < 0 && game.Answer() == ""
}`
	ctx := ixgo.NewContext(ixgo.EnableCachedReg | ixgo.SupportMultipleInterp)
	interp, err := ctx.LoadInterp("main.go", source)
	if err != nil {
		t.Fatal(err)
	}
	defer interp.UnsafeRelease()
	got, err := interp.RunFunc("Run")
	if err != nil || got != true {
		t.Fatalf("interpreted text and answer=%v, %v", got, err)
	}
}
