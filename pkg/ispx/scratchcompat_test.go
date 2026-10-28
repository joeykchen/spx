/*
 * Copyright (c) 2026 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ispx

import (
	"reflect"
	"testing"

	"github.com/goplus/ixgo"
	"github.com/goplus/ixgo/xgobuild"
	"github.com/goplus/spx/v4/pkg/scratchcompat"
)

const scratchCompatSource = `import (
	"math"
	"github.com/goplus/spx/v4"
	compat "github.com/goplus/spx/v4/pkg/scratchcompat"
)

func main() {
	if compat.CharAt("中文", 1) != "中" || compat.CharAt("中文", 2) != "文" {
		panic("one based character index")
	}
	if compat.CharAt("中文", 0) != "" || compat.CharAt("中文", -1) != "" || compat.CharAt("中文", 3) != "" {
		panic("invalid character index")
	}
	if spx.CharAt("中文", 0) != "中" || spx.CharAt("中文", 1) != "文" {
		panic("native character index must remain zero based")
	}
	number := spx.NewValue("03.5")
	items := spx.NewList("APPLE", number)
	if !compat.Equal(number, 3.5) || !compat.Equal(3.5, number) {
		panic("wrapped numeric equality")
	}
	if !compat.Equal(items.At(0), "apple") || !compat.Equal(items.At(1), number) {
		panic("list item equality")
	}
	if compat.Compare(number, 4) != -1 || compat.Compare("apple", items.At(0)) != 0 {
		panic("wrapped ordering")
	}
	if number.Raw() != "03.5" {
		panic("raw value")
	}
	near := math.Nextafter(math.Nextafter(1, 2), 2)
	if !compat.Equal(1.0, near) || compat.Equal(1.0, math.Nextafter(near, 2)) {
		panic("two ULP equality tolerance")
	}
	if compat.Compare(1.0, near) != -1 {
		panic("ordering must remain exact")
	}
	if compat.Div(0, 0) != 0 || compat.Div(math.NaN(), 1) != 0 {
		panic("division NaN normalization")
	}
	if !math.IsInf(compat.Div(1, 0), 1) || !math.IsInf(compat.Div(-1, 0), -1) {
		panic("division infinity preservation")
	}
	if compat.LimitPrecision(42 + 5e-10) != 42 || compat.LimitPrecision(42 + 2e-9) != 42 + 2e-9 {
		panic("coordinate precision threshold")
	}
	if !math.IsNaN(compat.LimitPrecision(math.NaN())) || !math.IsInf(compat.LimitPrecision(math.Inf(1)), 1) {
		panic("coordinate nonfinite preservation")
	}
	if compat.FloorMod(-7, 3) != 2 || compat.FloorMod(7, -3) != -2 {
		panic("floored modulo")
	}
	if !compat.Contains("Scratch", "RAT") || compat.Contains("Scratch", "dog") {
		panic("case insensitive contains")
	}
}
`

func TestInterpreterScratchCompat(t *testing.T) {
	for _, execution := range []string{"direct", "fallback"} {
		for _, optimization := range []string{"default", "optimized"} {
			t.Run(execution+"/"+optimization, func(t *testing.T) {
				isolateInterpreter(t)
				ctx := ixgo.NewContext(ixgo.SupportMultipleInterp | ixgo.EnableCachedReg | xgobuild.StaticLoad)
				called := make(map[string]bool)
				if execution == "fallback" {
					// Context overrides bypass generated direct-call adapters.
					for name, fn := range map[string]any{
						"CharAt":         scratchcompat.CharAt,
						"Compare":        scratchcompat.Compare,
						"Contains":       scratchcompat.Contains,
						"Div":            scratchcompat.Div,
						"Equal":          scratchcompat.Equal,
						"FloorMod":       scratchcompat.FloorMod,
						"LimitPrecision": scratchcompat.LimitPrecision,
					} {
						called[name] = false
						native := reflect.ValueOf(fn)
						wrapped := reflect.MakeFunc(native.Type(), func(args []reflect.Value) []reflect.Value {
							called[name] = true
							return native.Call(args)
						})
						ctx.RegisterExternal("github.com/goplus/spx/v4/pkg/scratchcompat."+name, wrapped.Interface())
					}
				}
				var options []InitOption
				if optimization == "optimized" {
					options = append(options, WithSSAOptimization())
				}
				if err := Init(ctx, options...); err != nil {
					t.Fatal(err)
				}
				if err := Build(map[string][]byte{"main.xgo": []byte(scratchCompatSource)}); err != nil {
					t.Fatal(err)
				}
				if code, err := Run(); code != 0 || err != nil {
					t.Fatalf("Run() = %d, %v", code, err)
				}
				for name, seen := range called {
					if !seen {
						t.Errorf("%s fallback was not called", name)
					}
				}
			})
		}
	}
}
