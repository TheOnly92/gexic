package main

import (
	"github.com/go-gl/gl"
	"math/rand"
	"time"
)

type HexMap [10][9]*Hex

type Hex struct {
	Kind  HexType
	State HexState
}

type HexType uint8
type HexState uint8

const (
	HEX_WIDTH  int = 44
	HEX_HEIGHT     = 40

	OddIncre = 19
)

const (
	HexEmpty HexType = iota

	HexRed
	HexGreen
	HexBlue
	HexYellow
	HexPurple
	HexLightBlue

	HexFlower
)

const (
	StateNormal HexState = iota
	StateRotating
	StateFalling
	StateRotatingStar
)

func GenHexMap() HexMap {
	rt := HexMap{}
	rand.Seed(time.Now().Unix())
	for x := 0; x < 10; x++ {
		maxy := 8
		if x%2 == 1 {
			maxy = 9
		}
		for y := 0; y < maxy; y++ {
			rt[x][y] = &Hex{HexType(rand.Intn(6) + 1), StateNormal}
		}
	}

	return rt
}

func (m HexMap) Render() {
	gl.Enable(gl.TEXTURE_2D)
	gl.Enable(gl.BLEND)
	gl.Disable(gl.DEPTH_TEST)
	for x := 0; x < 10; x++ {
		maxy := 8
		if x%2 == 1 {
			maxy = 9
		}
		for y := 0; y < maxy; y++ {
			if m[x][y].State != StateNormal {
				continue
			}
			gl.PushMatrix()
			posX, posY := m.GetTopLeft(x, y).WithOffset()
			gl.Translatef(float32(posX), float32(posY), 0)
			m[x][y].Render(1)
			gl.PopMatrix()
		}
	}
	gl.Flush()
	gl.Disable(gl.TEXTURE_2D)
	gl.Disable(gl.BLEND)
}

func (m HexMap) GetCenter(x, y int) Point {
	rt := m.GetTopLeft(x, y)
	rt.X += float64(HEX_WIDTH / 2)
	rt.Y += float64(HEX_HEIGHT / 2)
	return rt
}

func (m HexMap) GetTopLeft(x, y int) Point {
	rt := Point{}
	rt.X = float64(x * 33)
	rt.Y = float64(y * 38)
	if x%2 == 1 {
		rt.Y -= float64(OddIncre)
	}
	return rt
}

func (h *Hex) Render(alpha float32) {
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	if h.Kind == HexFlower {
		gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.REPLACE)
		starTex.Bind(gl.TEXTURE_2D)
	} else {
		gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
		hexTex.Bind(gl.TEXTURE_2D)
		var r, g, b float32
		switch h.Kind {
		case HexRed:
			r = 1
		case HexGreen:
			g = 1
		case HexBlue:
			b = 1
		case HexYellow:
			r = 1
			g = 1
		case HexPurple:
			r = 1
			b = 1
		case HexLightBlue:
			g = 1 - 222/255
			b = 1
		}
		if alpha < 1 {
			gl.Color4f(r, g, b, alpha)
		} else {
			gl.Color3f(r, g, b)
		}
	}
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 0)
	gl.Vertex2i(HEX_WIDTH/2, HEX_HEIGHT/2)
	gl.TexCoord2f(0, 1)
	gl.Vertex2i(HEX_WIDTH/2, -HEX_HEIGHT/2)
	gl.TexCoord2f(1, 1)
	gl.Vertex2i(-HEX_WIDTH/2, -HEX_HEIGHT/2)
	gl.TexCoord2f(1, 0)
	gl.Vertex2i(-HEX_WIDTH/2, HEX_HEIGHT/2)
	gl.End()
}
