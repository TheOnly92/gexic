package main

import (
	"fmt"
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

func (h *SelectedHex) String() string {
	return fmt.Sprintf("%x %x", h.Pos.X, h.Pos.Y)
}

func (r *AnimateRotate) InitAnimation(hexes []FieldPoint, timesToRotate int) {
	for _, p := range hexes {
		r.SelectedHex = append(r.SelectedHex, SelectedHex{hexMap2[p.X][p.Y], p})
		hexMap2[p.X][p.Y].State = StateRotating
	}
	r.TimesToRotate = timesToRotate
	r.RotateAngle = 0
	r.PauseTicks = 0
}

func (r *AnimateRotate) SetPostHook(f func()) {
	r.postHook = f
}

func (r *AnimateRotate) AnimateAndExecute() {
	if len(r.SelectedHex) == 0 {
		return
	}
	gl.PushMatrix()
	var p Point
	for _, hex := range r.SelectedHex {
		p.X += hexMap2.GetCenter(hex.Pos.X, hex.Pos.Y).X
		p.Y += hexMap2.GetCenter(hex.Pos.X, hex.Pos.Y).Y
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
		hex.Hex.Render(1, false)
		gl.PopMatrix()
	}
	gl.PopMatrix()
	if r.PauseTicks > 0 {
		return
	}
	if r.RotateAngle < 120 {
		r.RotateAngle += 15
	} else {
		fmt.Println(r.SelectedHex[0].Hex, r.SelectedHex[1].Hex, r.SelectedHex[2].Hex)
		hexMap2[r.SelectedHex[0].Pos.X][r.SelectedHex[0].Pos.Y], hexMap2[r.SelectedHex[1].Pos.X][r.SelectedHex[1].Pos.Y], hexMap2[r.SelectedHex[2].Pos.X][r.SelectedHex[2].Pos.Y] = hexMap2[r.SelectedHex[2].Pos.X][r.SelectedHex[2].Pos.Y], hexMap2[r.SelectedHex[0].Pos.X][r.SelectedHex[0].Pos.Y], hexMap2[r.SelectedHex[1].Pos.X][r.SelectedHex[1].Pos.Y]
		r.SelectedHex[0].Hex, r.SelectedHex[1].Hex, r.SelectedHex[2].Hex = r.SelectedHex[2].Hex, r.SelectedHex[0].Hex, r.SelectedHex[1].Hex
		collide := hexMap2.CheckCollision()
		if r.TimesToRotate == 0 || collide {
			for _, hex := range r.SelectedHex {
				if hex.Hex.State == StateRotating {
					hex.Hex.State = StateNormal
				}
			}
			r.SelectedHex = nil
			r.SelectedHex = make([]SelectedHex, 0)
			if collide {
				hexShrink.InitAnimation()
				if r.postHook != nil {
					hexShrink.postHook = r.postHook
				}
			} else {
				if r.postHook != nil {
					r.postHook()
				}
			}
		} else {
			r.TimesToRotate--
			r.RotateAngle = 0
			r.PauseTicks = 5
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
		hex.Hex.Render(1, false)
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

type ShrinkHex struct {
	SelectedHex []SelectedHex
	Scale       float32
	postHook    func()
}

func (s *ShrinkHex) InitAnimation() {
	for x := 0; x < 10; x++ {
		maxy := 8
		if x%2 == 1 {
			maxy = 9
		}
		for y := 0; y < maxy; y++ {
			if hexMap2[x][y].State == StateShrinking {
				s.SelectedHex = append(s.SelectedHex, SelectedHex{hexMap2[x][y], FieldPoint{x, y}})
			}
		}
	}
	s.Scale = 1
}

func (r *ShrinkHex) SetPostHook(f func()) {
	r.postHook = f
}

func (s *ShrinkHex) AnimateAndExecute() {
	if len(s.SelectedHex) == 0 {
		return
	}
	if s.Scale > 0 {
		s.Scale -= 0.1
		for _, hex := range s.SelectedHex {
			gl.PushMatrix()
			x, y := hexMap2.GetCenter(hex.Pos.X, hex.Pos.Y).WithOffset()
			gl.Translatef(float32(x), float32(y), 0)
			gl.Scalef(s.Scale, s.Scale, 1)
			hex.Hex.Render(1, true)
			gl.PopMatrix()
		}
	} else {
		if s.postHook != nil {
			s.postHook()
		}
		s.SelectedHex = nil
		s.SelectedHex = make([]SelectedHex, 0)
	}
}
