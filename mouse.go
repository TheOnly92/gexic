package main

import (
	// "fmt"
	"github.com/go-gl/glfw"
)

var mouse Mouse

var (
	hexRotate AnimateRotate
	hexShrink ShrinkHex
	hexFall   AnimateFall
)

type Mouse struct {
	pos          Point
	locked       bool
	selectedHex  []FieldPoint
	selectedStar bool
}

func (m *Mouse) GetXY() (int, int) {
	return int(m.pos.X), int(m.pos.Y)
}

func MousePosCallback(x, y int) {
	if mouse.locked {
		return
	}
	mouse.pos.X = float64(x)
	mouse.pos.Y = float64(y)
}

func MouseButtonCallback(button, state int) {
	if mouse.locked {
		return
	}
	if state == glfw.KeyPress {
		switch button {
		case glfw.MouseLeft:
			if mouse.selectedStar {

			} else if len(mouse.selectedHex) == 3 {
				mouse.locked = true
				hexRotate.InitAnimation(mouse.selectedHex, 2)
				hexRotate.SetPostHook(func() {
					mouse.locked = false
				})
			}
		}
	}
}
