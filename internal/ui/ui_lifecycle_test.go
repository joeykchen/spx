//go:build !pure_engine

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

package ui

import (
	"slices"
	"testing"

	"github.com/goplus/spbase/mathf"
	coreproject "github.com/goplus/spx/v4/internal/core/project"
	"github.com/goplus/spx/v4/internal/coroutine"
	"github.com/goplus/spx/v4/internal/engine"
	gdx "github.com/goplus/spx/v4/pkg/spx/pkg/engine"
)

type uiLifecycleBinding struct {
	parent gdx.Object
	path   string
}

type uiLifecycleMgr struct {
	*uiRenderSpy
	next     gdx.Object
	prefabs  map[gdx.Object]string
	bindings map[gdx.Object]uiLifecycleBinding
	monitor  *monitorRenderSpy
}

type uiLifecyclePlatform struct{ gdx.IPlatformMgr }

func TestUIPrefabsBindChildrenBeforeFirstUse(t *testing.T) {
	for _, test := range []struct {
		name   string
		create func() gdx.IUiNode
	}{
		{"UiAsk", func() gdx.IUiNode { return engine.CreateEngineUIForType[UiAsk]("") }},
		{"UiDebug", func() gdx.IUiNode { return engine.CreateEngineUIForType[UiDebug]("") }},
		{"UiMeasure", func() gdx.IUiNode { return engine.CreateEngineUIForType[UiMeasure]("") }},
		{"UiQuote", func() gdx.IUiNode { return engine.CreateEngineUIForType[UiQuote]("") }},
		{"UiSay", func() gdx.IUiNode { return engine.CreateEngineUIForType[UiSay]("") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			mgr := setupUILifecycle(t)
			node := test.create()
			if got, want := mgr.prefabs[node.GetId()], "res://engine/ui/"+test.name+".tscn"; got != want {
				t.Fatalf("loaded prefab = %q, want %q", got, want)
			}
			if gdx.UiNodes()[node.GetId()] != node {
				t.Fatal("constructor did not register the live UI instance")
			}
			var children map[string]*UiNode
			switch panel := node.(type) {
			case *UiAsk:
				children = map[string]*UiNode{
					"MF/Frame/AskBody": panel.askBody, "MF/Frame/AskBody/LabelAsk": panel.askLabel,
					"M/Input": panel.input, "M/Input/Check": panel.checkBtn,
				}
			case *UiDebug:
				children = map[string]*UiNode{"Label": panel.input}
			case *UiMeasure:
				children = map[string]*UiNode{"C": panel.container, "C/Line": panel.imageLine, "LC": panel.labelContainer, "LC/Label": panel.labelValue}
			case *UiQuote:
				children = map[string]*UiNode{
					"C": panel.container, "C/ImageL": panel.imageL, "C/ImageR": panel.imageR,
					"C/LabelDes": panel.labelDes, "C/LabelMsg": panel.labelMsg,
				}
			case *UiSay:
				children = map[string]*UiNode{
					"VL": panel.left.vbox, "VL/BG/Label": panel.left.label,
					"VR": panel.right.vbox, "VR/BG/Label": panel.right.label,
					"VLThink": panel.leftThink.vbox, "VLThink/BG/MC/Label": panel.leftThink.label,
					"VRThink": panel.rightThink.vbox, "VRThink/BG/MC/Label": panel.rightThink.label,
				}
				// No cloud may flash at the origin before its first layout.
				for _, variant := range []sayNodes{panel.left, panel.right, panel.leftThink, panel.rightThink} {
					visible, set := mgr.visible[variant.vbox.GetId()]
					if !set || visible {
						t.Fatal("new bubble did not explicitly hide every variant")
					}
				}
			}
			if len(mgr.bindings) != len(children) {
				t.Fatalf("bound %d nodes, want %d", len(mgr.bindings), len(children))
			}
			for path, child := range children {
				if child == nil {
					t.Fatalf("child %q was not bound", path)
				}
				if got, want := mgr.bindings[child.GetId()], (uiLifecycleBinding{node.GetId(), path}); got != want {
					t.Errorf("child binding = %+v, want %+v", got, want)
				}
				if gdx.UiNodes()[child.GetId()] != child || child.OnUiClickEvent == nil {
					t.Errorf("child %q is not fully initialized", path)
				}
			}
		})
	}
}

func TestUiMonitorFactoryInitializesFirstRender(t *testing.T) {
	for appearance := MonitorAppearanceDefault; appearance < monitorAppearanceCount; appearance++ {
		spec := monitorViewSpecs[appearance]
		t.Run(spec.root, func(t *testing.T) {
			mgr := setupUILifecycle(t)
			engine.SetWindowScale(1)
			style := MonitorStyle{
				Appearance: appearance, Label: "score", Color: mathf.NewColorRGBAi(255, 128, 64, 255),
				Dimensions: mathf.NewVec2(140, 240), Slider: MonitorSlider{-1, 100, 0.01},
			}
			panel := NewUiMonitor(style)
			if mgr.prefabs[panel.GetId()] != "res://engine/ui/UiMonitor.tscn" || gdx.UiNodes()[panel.GetId()] != panel {
				t.Fatal("monitor factory did not create and register its prefab")
			}
			if _, changed := panel.ReadSliderChange(); changed {
				t.Fatal("constructor reported slider input before the first render")
			}
			bound := make(map[string]gdx.Object)
			for id, binding := range mgr.bindings {
				if binding.parent != panel.GetId() || gdx.UiNodes()[id] == nil {
					t.Fatalf("unregistered monitor child: %+v", binding)
				}
				bound[binding.path] = id
			}
			wantBindings := 0
			for _, path := range []string{spec.root, spec.label, spec.value, spec.colorTarget, spec.input} {
				if path != "" {
					wantBindings++
					if bound[path] == 0 {
						t.Fatalf("monitor child %q was not bound", path)
					}
				}
			}
			if len(bound) != wantBindings || !mgr.visible[bound[spec.root]] {
				t.Fatalf("selected view bindings/visibility = %v/%v", bound, mgr.visible)
			}
			if spec.label != "" && mgr.text[bound[spec.label]] != style.Label || spec.colorTarget != "" && mgr.colors[bound[spec.colorTarget]] != style.Color {
				t.Fatal("constructor lost the fixed label or color")
			}
			panel.UpdateScale(1.5)
			panel.UpdatePos(mathf.NewVec2(10, 20))
			assertVec2Near(t, "monitor scale", mgr.scales[panel.GetId()], mathf.NewVec2(1.5, 1.5))
			assertVec2Near(t, "monitor position", mgr.positions[panel.GetId()], mathf.NewVec2(240, 140))
			value := MonitorValue{Text: "42", Items: []string{"first", "second"}}
			panel.Render(value)
			if appearance == MonitorAppearanceList {
				id := bound[spec.root]
				if mgr.sizes[id] != style.Dimensions || mgr.monitor.text[id] != style.Label || mgr.monitor.color[id] != style.Color || !slices.Equal(mgr.monitor.items[id], value.Items) {
					t.Fatal("list factory lost style or initial items")
				}
				value.Items[0] = "changed"
				panel.Render(value)
				if !slices.Equal(mgr.monitor.items[id], value.Items) || mgr.monitor.listCalls != 3 {
					t.Fatal("list refresh missed the changed source items")
				}
			} else if mgr.text[bound[spec.value]] != value.Text {
				t.Fatal("first monitor render lost its value")
			}
			if appearance == MonitorAppearanceSlider {
				if _, changed := panel.ReadSliderChange(); changed {
					t.Fatal("rendered slider value was reported as user input")
				}
				mgr.monitor.ranges[bound[spec.input]] = 43
				if value, changed := panel.ReadSliderChange(); !changed || value != 43 {
					t.Fatalf("slider input = %v/%v, want 43/true", value, changed)
				}
			}
			for _, visible := range []bool{true, false, true} {
				panel.SetVisible(visible)
				if mgr.visible[panel.GetId()] != visible || !mgr.visible[bound[spec.root]] {
					t.Fatal("visibility update changed the selected view")
				}
			}
		})
	}
}

func TestUIAskBoundButtonSubmitsCurrentAnswer(t *testing.T) {
	mgr := setupUILifecycle(t)
	panel := engine.CreateEngineUIForType[UiAsk]("")
	for _, answer := range []string{"first answer", "下一题"} {
		calls := 0
		panel.Show(true, "Question", func(got string) {
			calls++
			if got != answer || mgr.visible[panel.GetId()] {
				t.Errorf("submitted %q with visible=%v", got, mgr.visible[panel.GetId()])
			}
		})
		mgr.text[panel.input.GetId()] = answer
		panel.checkBtn.V_OnUiClick()
		if calls != 1 {
			t.Fatalf("button invoked callback %d times, want once", calls)
		}
	}
}

func TestMeasureShapeKeepsLabelInLogicalUnits(t *testing.T) {
	for _, test := range []struct {
		name, text             string
		size, scale, lineWidth float64
	}{
		{"integer", "10", 10, 3, 38},
		{"rounded", "1.3", 1.26, 2, 10.52},
	} {
		t.Run(test.name, func(t *testing.T) {
			mgr := setupUILifecycle(t)
			color := mathf.NewColorRGBAi(20, 40, 60, 255)
			shape := NewMeasureShape(coreproject.MeasureShape{X: 5, Y: 7, Size: test.size, Scale: test.scale, Heading: 90, Color: color}).(*measureShape)
			panel := shape.panel
			if got := mgr.text[panel.labelValue.GetId()]; got != test.text {
				t.Fatalf("measurement label = %q, want %q", got, test.text)
			}
			assertVec2Near(t, "scaled ruler", mgr.sizes[panel.container.GetId()], mathf.NewVec2(test.lineWidth, 26))
			assertVec2Near(t, "label anchor", mgr.positions[panel.labelContainer.GetId()], mathf.NewVec2(240, 146))
			if mgr.colors[panel.container.GetId()] != color || mgr.colors[panel.labelContainer.GetId()] != color {
				t.Fatal("measurement lost its configured color")
			}
		})
	}
}

func (m *uiLifecycleMgr) CreateNode(path string) gdx.Object {
	m.next++
	m.prefabs[m.next] = path
	return m.next
}

func (m *uiLifecycleMgr) BindNode(parent gdx.Object, path string) gdx.Object {
	m.next++
	m.bindings[m.next] = uiLifecycleBinding{parent, path}
	return m.next
}

func (m *uiLifecycleMgr) SetListItems(id gdx.Object, name string, items engine.Array, color mathf.Color) {
	m.monitor.SetListItems(id, name, items, color)
}

func (m *uiLifecycleMgr) SetRange(id gdx.Object, min, max, step, value float64) {
	m.monitor.SetRange(id, min, max, step, value)
}

func (m *uiLifecycleMgr) GetRangeValue(id gdx.Object) float64 { return m.monitor.GetRangeValue(id) }

func (*uiLifecycleMgr) GetCanvasScale(gdx.Object) mathf.Vec2 { return mathf.NewVec2(1, 1) }

func (m *uiLifecycleMgr) SetTextMaxWidth(id gdx.Object, text string, width float64) float64 {
	m.text[id] = text
	return m.monitor.SetTextMaxWidth(id, text, width)
}

func (uiLifecyclePlatform) IsMainThread() bool { return true }

func setupUILifecycle(t *testing.T) *uiLifecycleMgr {
	t.Helper()
	mgr := &uiLifecycleMgr{
		uiRenderSpy: setupUIRendering(t), next: 100000, monitor: newMonitorRenderSpy(),
		prefabs: make(map[gdx.Object]string), bindings: make(map[gdx.Object]uiLifecycleBinding),
	}
	for id := range gdx.UiNodes() {
		mgr.next = max(mgr.next, id)
	}
	gdx.UiMgr = mgr
	// Public factories cross the engine-thread boundary. The UI package has
	// no running game; install an idle scheduler for them.
	previous := gdx.PlatformMgr
	engine.SetCoroutines(coroutine.New(nil))
	gdx.PlatformMgr = uiLifecyclePlatform{}
	t.Cleanup(func() {
		engine.SetCoroutines(nil)
		gdx.PlatformMgr = previous
		for id := range mgr.prefabs {
			delete(gdx.UiNodes(), id)
		}
		for id := range mgr.bindings {
			delete(gdx.UiNodes(), id)
		}
	})
	return mgr
}
