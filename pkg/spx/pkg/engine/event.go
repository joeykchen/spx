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

package engine

import (
	"maps"
	"slices"
	"sync"
)

type Action0 func()

type Event0 struct {
	listeners eventListeners[Action0]
}

type Action1[T any] func(data T)

type Event1[T any] struct {
	listeners eventListeners[Action1[T]]
}

type Action2[T1 any, T2 any] func(data1 T1, data2 T2)

type Event2[T1 any, T2 any] struct {
	listeners eventListeners[Action2[T1, T2]]
}

type eventListeners[A any] struct {
	actions map[int]A
	nextID  int
	mutex   sync.Mutex
}

func NewEvent0() *Event0 {
	return &Event0{}
}

func NewEvent1[T any]() *Event1[T] {
	return &Event1[T]{}
}

func NewEvent2[T1 any, T2 any]() *Event2[T1, T2] {
	return &Event2[T1, T2]{}
}

func (e *Event0) Subscribe(action Action0) int {
	return e.listeners.subscribe(action)
}

func (e *Event0) Unsubscribe(id int) {
	e.listeners.unsubscribe(id)
}

func (e *Event0) UnsubscribeAll() {
	e.listeners.unsubscribeAll()
}

func (e *Event0) Trigger() {
	for _, action := range e.listeners.snapshot() {
		action()
	}
}

func (e *Event1[T]) Subscribe(action Action1[T]) int {
	return e.listeners.subscribe(action)
}

func (e *Event1[T]) Unsubscribe(id int) {
	e.listeners.unsubscribe(id)
}

func (e *Event1[T]) UnsubscribeAll() {
	e.listeners.unsubscribeAll()
}

func (e *Event1[T]) Trigger(data T) {
	for _, action := range e.listeners.snapshot() {
		action(data)
	}
}

func (e *Event2[T1, T2]) Subscribe(action Action2[T1, T2]) int {
	return e.listeners.subscribe(action)
}

func (e *Event2[T1, T2]) Unsubscribe(id int) {
	e.listeners.unsubscribe(id)
}

func (e *Event2[T1, T2]) UnsubscribeAll() {
	e.listeners.unsubscribeAll()
}

func (e *Event2[T1, T2]) Trigger(data1 T1, data2 T2) {
	for _, action := range e.listeners.snapshot() {
		action(data1, data2)
	}
}

func (e *eventListeners[A]) subscribe(action A) int {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.actions == nil {
		e.actions = make(map[int]A)
	}
	id := e.nextID
	e.actions[id] = action
	e.nextID++

	return id
}

func (e *eventListeners[A]) unsubscribe(id int) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	delete(e.actions, id)
}

func (e *eventListeners[A]) unsubscribeAll() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.actions = nil
}

// Each dispatch owns its callbacks, so nested or concurrent triggers cannot
// change an active snapshot. Callbacks run after snapshot releases the mutex.
func (e *eventListeners[A]) snapshot() []A {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	ids := slices.AppendSeq(make([]int, 0, len(e.actions)), maps.Keys(e.actions))
	slices.Sort(ids)
	actions := make([]A, len(ids))
	for i, id := range ids {
		actions[i] = e.actions[id]
	}
	return actions
}
