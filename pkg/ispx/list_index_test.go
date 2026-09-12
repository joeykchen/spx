package ispx

import (
	"github.com/goplus/ixgo"
	"testing"
)

func TestInterpretedScratchListIndices(t *testing.T) {
	const source = `package main
import "github.com/goplus/spx/v3"
func Run() string {
 list := spx.NewList("a", "b")
 index := spx.NewValue(-2)
 list.Delete(spx.ScratchListIndex(index))
 list.Insert(spx.ScratchListIndex(-2), "x")
 list.Insert(spx.ScratchListIndex(999), "x")
 list.Set(spx.ScratchListIndex("last"), "c")
 return list.String()
}`
	ctx := ixgo.NewContext(ixgo.EnableCachedReg | ixgo.SupportMultipleInterp)
	interp, err := ctx.LoadInterp("main.go", source)
	if err != nil {
		t.Fatal(err)
	}
	defer interp.UnsafeRelease()
	got, err := interp.RunFunc("Run")
	if err != nil || got != "ac" {
		t.Fatalf("interpreted list operations = %v, %v", got, err)
	}
}
