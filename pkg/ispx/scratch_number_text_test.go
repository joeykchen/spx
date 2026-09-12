package ispx

import (
	"github.com/goplus/ixgo"
	"testing"
)

func TestInterpretedScratchNumberText(t *testing.T) {
	const source = `package main
import "github.com/goplus/spx/v3"
func Run() bool {
var number spx.Value
number.Set(1e6)
list := spx.NewList(1e-7, "x")
return number.String() == "1000000" && list.String() == "1e-7 x"
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
