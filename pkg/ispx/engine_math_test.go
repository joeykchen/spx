package ispx

import (
	"testing"

	"github.com/goplus/ixgo"
)

func TestInterpretedEngineAngleMath(t *testing.T) {
	const source = `package main
import "github.com/goplus/spx/v3/pkg/spx/pkg/engine"
func main() {
	if engine.NormalizeDegrees(-900) != 180 {
		panic("direction normalization")
	}
	from, to := engine.NormalizeAngleRange(-1079, 719)
	if to-from != -2 {
		panic("shortest rotation path")
	}
}`
	if _, err := ixgo.NewContext(0).RunFile("main.go", source, nil); err != nil {
		t.Fatal(err)
	}
}
