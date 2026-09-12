package spx

const scratchMissingTargetDistance = 10000

// resolveTargetPosition distinguishes a missing target from one at the origin.
func (p *Game) resolveTargetPosition(obj Target) (float64, float64, bool) {
	var sprite *SpriteImpl
	switch v := obj.(type) {
	case SpriteName:
		sprite = p.findSprite(v)
	case Sprite:
		sprite = spriteOf(v)
	case specialObj:
		if v == Mouse {
			x, y := p.getMousePos()
			return x, y, true
		}
	case Pos:
		if v == Random {
			worldW, worldH := p.worldSize()
			mx, my := randomIntn(worldW), randomIntn(worldH)
			return float64(mx - (worldW >> 1)), float64((worldH >> 1) - my), true
		}
	}
	if sprite == nil {
		return 0, 0, false
	}
	x, y := sprite.getXY()
	return x, y, true
}
