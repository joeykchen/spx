/*
 * Copyright (c) 2021 The XGo Authors (xgo.dev). All rights reserved.
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

// export by github.com/goplus/ixgo/cmd/qexp

package scratchcompat

import (
	q "github.com/goplus/spx/v4/pkg/scratchcompat"

	"reflect"

	"github.com/goplus/ixgo"
)

func init() {
	ixgo.RegisterPackage(&ixgo.Package{
		Name: "scratchcompat",
		Path: "github.com/goplus/spx/v4/pkg/scratchcompat",
		Deps: map[string]string{
			"github.com/goplus/spx/v4":                  "spx",
			"github.com/goplus/spx/v4/internal/scratch": "scratch",
		},
		Interfaces: map[string]reflect.Type{},
		NamedTypes: map[string]reflect.Type{},
		AliasTypes: map[string]reflect.Type{},
		Vars:       map[string]reflect.Value{},
		Funcs: map[string]reflect.Value{
			"CharAt":         reflect.ValueOf(q.CharAt),
			"Compare":        reflect.ValueOf(q.Compare),
			"Contains":       reflect.ValueOf(q.Contains),
			"Div":            reflect.ValueOf(q.Div),
			"Equal":          reflect.ValueOf(q.Equal),
			"FloorMod":       reflect.ValueOf(q.FloorMod),
			"LimitPrecision": reflect.ValueOf(q.LimitPrecision),
		},
		TypedConsts:   map[string]ixgo.TypedConst{},
		UntypedConsts: map[string]ixgo.UntypedConst{},
	})
	ixgo.RegisterDirectCalls("github.com/goplus/spx/v4/pkg/scratchcompat", map[string]ixgo.DirectCallAdapter{
		"CharAt":         func_CharAt,
		"Compare":        func_Compare,
		"Contains":       func_Contains,
		"Div":            func_Div,
		"Equal":          func_Equal,
		"FloorMod":       func_FloorMod,
		"LimitPrecision": func_LimitPrecision,
	})
}

func func_CharAt(ctx ixgo.DirectCallContext) {
	ctx.SetResult(q.CharAt(ixgo.DirectCallArg[string](ctx, 0), ixgo.DirectCallArg[int](ctx, 1)))
}

func func_Compare(ctx ixgo.DirectCallContext) {
	ctx.SetResult(q.Compare(ixgo.DirectCallArg[any](ctx, 0), ixgo.DirectCallArg[any](ctx, 1)))
}

func func_Contains(ctx ixgo.DirectCallContext) {
	ctx.SetResult(q.Contains(ixgo.DirectCallArg[string](ctx, 0), ixgo.DirectCallArg[string](ctx, 1)))
}

func func_Div(ctx ixgo.DirectCallContext) {
	ctx.SetResult(q.Div(ixgo.DirectCallArg[float64](ctx, 0), ixgo.DirectCallArg[float64](ctx, 1)))
}

func func_Equal(ctx ixgo.DirectCallContext) {
	ctx.SetResult(q.Equal(ixgo.DirectCallArg[any](ctx, 0), ixgo.DirectCallArg[any](ctx, 1)))
}

func func_FloorMod(ctx ixgo.DirectCallContext) {
	ctx.SetResult(q.FloorMod(ixgo.DirectCallArg[float64](ctx, 0), ixgo.DirectCallArg[float64](ctx, 1)))
}

func func_LimitPrecision(ctx ixgo.DirectCallContext) {
	ctx.SetResult(q.LimitPrecision(ixgo.DirectCallArg[float64](ctx, 0)))
}
