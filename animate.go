package main

import (
	"github.com/go-gl/gl"
	"math"
)

type AnimateRotate struct {
	SelectedHex   []SelectedHex
	RotateAngle   float32
	TimesToRotate int
	PauseTicks    int
	postHook      func()
}

type SelectedHex struct {
	Hex *Hex
	Pos FieldPoint
}

func (r *AnimateRotate) InitAnimation(hexes []FieldPoint, timesToRotate int) {
	for _, p := range hexes {
		r.SelectedHex = append(r.SelectedHex, SelectedHex{hexMap2[p.X][p.Y], p})
		hexMap2[p.X][p.Y].State = StateRotating
	}
	r.TimesToRotate = timesToRotate
}

func (r *AnimateRotate) SetPostHook(f func()) {
	r.postHook = f
}

func (r *AnimateRotate) AnimateAndExecute() {
	if r.TimesToRotate == 0 {
		return
	}
	gl.PushMatrix()
	var p Point
	for _, hex := range r.SelectedHex {
		p.X += hexMap2.GetTopLeft(hex.Pos.X, hex.Pos.Y).X
		p.Y += hexMap2.GetTopLeft(hex.Pos.X, hex.Pos.Y).Y
	}
	p.X /= 3
	p.Y /= 3
	x, y := p.WithOffset()
	gl.Translatef(float32(x), float32(y), 0)
	if r.PauseTicks == 0 {
		gl.Scalef(1.3, 1.3, 1)
		gl.Rotatef(r.RotateAngle, 0, 0, 1)
	} else {
		r.PauseTicks--
	}
	for _, hex := range r.SelectedHex {
		gl.PushMatrix()
		x2, y2 := hexMap2.GetTopLeft(hex.Pos.X, hex.Pos.Y).WithOffset()
		gl.Translatef(float32(x2-x), float32(y2-y), 0)
		hex.Hex.Render(1)
		gl.PopMatrix()
	}
	gl.PopMatrix()
	if r.PauseTicks > 0 {
		return
	}
	if r.RotateAngle < 120 {
		r.RotateAngle += 15
	} else {
		hexMap2[r.SelectedHex[0].Pos.X][r.SelectedHex[0].Pos.Y], hexMap2[r.SelectedHex[1].Pos.X][r.SelectedHex[1].Pos.Y], hexMap2[r.SelectedHex[2].Pos.X][r.SelectedHex[2].Pos.Y] = hexMap2[r.SelectedHex[2].Pos.X][r.SelectedHex[2].Pos.Y], hexMap2[r.SelectedHex[0].Pos.X][r.SelectedHex[0].Pos.Y], hexMap2[r.SelectedHex[1].Pos.X][r.SelectedHex[1].Pos.Y]
		r.TimesToRotate--
		r.RotateAngle = 0
		r.PauseTicks = 5
		if r.TimesToRotate == 0 {
			for _, hex := range r.SelectedHex {
				hex.Hex.State = StateNormal
			}
			r.SelectedHex = nil
			r.SelectedHex = make([]SelectedHex, 0)
		}
		if r.postHook != nil {
			r.postHook()
		}
	}
}

type AnimateFall struct {
	FallHex   []FallHex
	FallTicks float64
	postHook  func()
}

type FallHex struct {
	Hex    *Hex
	Target FieldPoint
	Pos    FieldPoint
	Accel  float64
}

func (f *AnimateFall) InitAnimation(fallHexes []FallHex) {
	for _, f := range fallHexes {
		f.Accel = math.Pow(float64(f.Pos.Y), 2)/16 + 1
	}
	f.FallHex = fallHexes
	f.FallTicks = 0
}

func (r *AnimateFall) SetPostHook(f func()) {
	r.postHook = f
}

func (f *AnimateFall) AnimateAndExecute() {
	if len(f.FallHex) == 0 {
		return
	}
	stillFalling := 0
	for _, hex := range f.FallHex {
		gl.PushMatrix()
		x, y := hexMap2.GetTopLeft(hex.Pos.X, hex.Pos.Y).WithOffset()
		displaceY := hex.Accel * math.Pow(f.FallTicks, 2) / 2
		_, tY := hexMap2.GetTopLeft(hex.Pos.X, hex.Target.Y).WithOffset()
		newY := math.Min(y+displaceY, tY)
		gl.Translatef(float32(x), float32(newY), 0)
		hex.Hex.Render(1)
		gl.PopMatrix()
		if newY < tY {
			stillFalling++
		}
	}
	f.FallTicks++
	if stillFalling == 0 {
		f.FallHex = nil
		f.FallHex = make([]FallHex, 0)
		f.postHook()
	}
}
