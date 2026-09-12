package audio

import (
	"math"

	"github.com/goplus/spx/v3/internal/engine"
)

const scratchPitchStepsPerOctave = 120

func (m *Manager) GetPan(soundObj engine.Object) float64 {
	return m.backend.GetPan(soundObj) * 100
}

func (m *Manager) SetPan(soundObj engine.Object, value float64) {
	value = clampSoundControl(value, -100, 100)
	m.backend.SetPan(soundObj, value/100)
}

func (m *Manager) ChangePan(soundObj engine.Object, delta float64) {
	delta = soundControlNumber(delta)
	m.SetPan(soundObj, m.GetPan(soundObj)+delta)
}

func (m *Manager) GetPitch(soundObj engine.Object) float64 {
	return pitchScaleToScratchEffect(m.backend.GetPitch(soundObj))
}

func (m *Manager) SetPitch(soundObj engine.Object, value float64) {
	value = clampSoundControl(value, -360, 360)
	m.backend.SetPitch(soundObj, scratchPitchEffectToScale(value))
}

func (m *Manager) ChangePitch(soundObj engine.Object, delta float64) {
	delta = soundControlNumber(delta)
	m.SetPitch(soundObj, m.GetPitch(soundObj)+delta)
}

func (m *Manager) GetVolume(soundObj engine.Object) float64 {
	return m.backend.GetVolume(soundObj) * 100
}

func (m *Manager) SetVolume(soundObj engine.Object, value float64) {
	value = clampSoundControl(value, 0, 100)
	m.backend.SetVolume(soundObj, value/100)
}

func (m *Manager) ChangeVolume(soundObj engine.Object, delta float64) {
	delta = soundControlNumber(delta)
	m.SetVolume(soundObj, m.GetVolume(soundObj)+delta)
}

func scratchPitchEffectToScale(value float64) float64 {
	return math.Pow(2, value/scratchPitchStepsPerOctave)
}

func pitchScaleToScratchEffect(scale float64) float64 {
	if scale <= 0 {
		return 0
	}
	return scratchPitchStepsPerOctave * math.Log2(scale)
}

func soundControlNumber(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	return value
}

func clampSoundControl(value, low, high float64) float64 {
	return math.Min(math.Max(soundControlNumber(value), low), high)
}
