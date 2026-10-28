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

package spx

import (
	"strings"

	"github.com/goplus/spx/v4/internal/scratch"
)

// -----------------------------------------------------------------------------
// Positions
// -----------------------------------------------------------------------------
type Pos = int

const (
	Invalid Pos = -1
	Last    Pos = -2
	All         = -3 // Pos or StopKind
	Random  Pos = -4
)

// -----------------------------------------------------------------------------
// Values
// -----------------------------------------------------------------------------
type obj = any

// -----------------------------------------------------------------------------
// Value
// -----------------------------------------------------------------------------
type Value struct {
	data any
}

// -----------------------------------------------------------------------------
// List
// -----------------------------------------------------------------------------
type List struct {
	data []obj
}

// NewValue creates a new Value initialized with the given object.
func NewValue(o obj) Value {
	return Value{data: unwrapValue(o)}
}

// NewList creates a new List initialized with the given elements.
func NewList(l ...obj) List {
	data := make([]obj, len(l))
	for i, v := range l {
		data[i] = unwrapValue(v)
	}
	return List{data: data}
}

// Raw returns the stored value without converting its type or contents.
func (p Value) Raw() any {
	return p.data
}

func (p Value) Equal(v obj) bool {
	return Equal(p.data, v)
}

func (p Value) String() string {
	return scratch.ToString(p.data)
}

func (p Value) Int() int {
	i, _ := scratch.ToInt(p.data)
	return i
}

func (p Value) Float() float64 {
	f, _ := scratch.ToFloat64(p.data)
	return f
}

func (p *Value) Set(v obj) {
	if p == nil {
		return
	}
	p.data = unwrapValue(v)
}

func (p *List) Init(data ...obj) {
	p.data = data
}

func (p *List) InitFrom(src *List) {
	data := make([]obj, len(src.data))
	copy(data, src.data)
	p.data = data
}

func (p *List) Len() int {
	return len(p.data)
}

func (p *List) String() string {
	sep := ""
	items := make([]string, len(p.data))
	for i, item := range p.data {
		items[i] = scratch.ToString(item)
		if len(items[i]) != 1 {
			sep = " "
		}
	}
	return strings.Join(items, sep)
}

// Contains returns true if the list contains the element v.
func (p *List) Contains(v obj) bool {
	return p.IndexOf(v) != Invalid
}

// Append adds the element v to the end of the list.
func (p *List) Append(v obj) {
	p.data = append(p.data, unwrapValue(v))
}

// Set sets the element at the specified index i to v.
func (p *List) Set(i Pos, v obj) {
	n := len(p.data)
	if i < 0 {
		i = Pos(getListPos(i, n))
	}
	if i >= 0 && int(i) < n {
		p.data[i] = unwrapValue(v)
	}
}

// Insert inserts the element v at the specified index i.
func (p *List) Insert(i Pos, v obj) {
	n := len(p.data)
	if i < 0 {
		if i == Invalid {
			return
		}
		i = Pos(getListPos(i, n+1))
	}
	val := unwrapValue(v)
	p.data = append(p.data, val)
	if int(i) < n {
		copy(p.data[i+1:], p.data[i:])
		p.data[i] = val
	}
}

// Delete removes the element at the specified index.
func (p *List) Delete(i Pos) {
	n := len(p.data)
	if i < 0 {
		if i == All {
			p.data = p.data[:0]
			return
		}
		i = Pos(getListPos(i, n))
	}
	if i >= 0 && int(i) < n {
		p.data = append(p.data[:i], p.data[i+1:]...)
	}
}

// At returns the Value at the specified index.
func (p *List) At(i Pos) Value {
	n := len(p.data)
	if i < 0 {
		i = Pos(getListPos(i, n))
	}
	if i < 0 || int(i) >= n {
		return Value{}
	}
	return Value{p.data[i]}
}

// IndexOf returns the zero-based position of the first occurrence of v in the list.
// Returns Invalid (-1) if v is not found.
func (p *List) IndexOf(v obj) Pos {
	for i, item := range p.data {
		if Equal(item, v) {
			return Pos(i)
		}
	}
	return Invalid
}

// Clear removes all elements from the list.
func (p *List) Clear() {
	p.data = p.data[:0]
}

func unwrapValue(v obj) any {
	if o, ok := v.(Value); ok {
		return o.data
	}
	return v
}

func getListPos(i Pos, n int) int {
	if i == Last {
		return n - 1
	}
	if i == Random {
		if n == 0 {
			return 0
		}
		return int(randomInt31n(int32(n)))
	}
	return int(i)
}
