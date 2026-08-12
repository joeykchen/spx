//go:build !js

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

package platform

import (
	"testing"

	gdx "github.com/goplus/spx/v3/pkg/spx/pkg/engine"
)

type testPlatformMgr struct {
	gdx.IPlatformMgr
	main bool
}

func (p testPlatformMgr) IsMainThread() bool {
	return p.main
}

func TestCanCallEngineDirectlyUsesGodot(t *testing.T) {
	previous := gdx.PlatformMgr
	t.Cleanup(func() { gdx.PlatformMgr = previous })

	gdx.PlatformMgr = testPlatformMgr{main: true}
	if !CanCallEngineDirectly() {
		t.Fatal("Godot main thread was not reported")
	}

	gdx.PlatformMgr = testPlatformMgr{main: false}
	if CanCallEngineDirectly() {
		t.Fatal("Godot worker thread was reported as the main thread")
	}
}
