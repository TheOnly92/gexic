package main

import (
	"encoding/json"
	"github.com/go-gl/gl"
	"io/ioutil"
	"math/rand"
	"time"
)

type Colors struct {
	Colors [][]int `json:"colors"`
}

var colors Colors

func init() {
	c, err := ioutil.ReadFile("colors.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(c, &colors)
	if err != nil {
		panic(err)
	}
}

type HexMap [11][9]*Hex

type Hex struct {
	Kind  HexType
	State HexState
}

type HexType uint8
type HexState uint8

const (
	HEX_WIDTH  int = 106
	HEX_HEIGHT     = 106

	OddIncre = 45
)

const (
	HexEmpty HexType = iota

	HexRed
	HexGreen
	HexBlue
	HexOrange
	HexPurple
	HexCyan

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
	for x := 0; x < 11; x++ {
		maxy := 8
		if x%2 == 1 {
			maxy = 9
		}
		for y := 0; y < maxy; y++ {
			rt[x][y] = &Hex{HexType(rand.Intn(int(HexFlower)) + 1), StateNormal}
		}
	}

	return rt
}

func (m HexMap) Render() {
	for x := 0; x < 11; x++ {
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
}

func (m HexMap) GetCenter(x, y int) Point {
	rt := m.GetTopLeft(x, y)
	rt.X += float64(HEX_WIDTH / 2)
	rt.Y += float64(HEX_HEIGHT / 2)
	return rt
}

func (m HexMap) GetTopLeft(x, y int) Point {
	rt := Point{}
	rt.X = float64(x * 77)
	rt.Y = float64(y * 90)
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
		gl.GetError()
		var r, g, b uint8
		r = uint8(colors.Colors[h.Kind-1][0])
		g = uint8(colors.Colors[h.Kind-1][1])
		b = uint8(colors.Colors[h.Kind-1][2])
		if alpha < 1 {
			gl.Color4ub(r, g, b, uint8(alpha*255))
		} else {
			gl.Color3ub(r, g, b)
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
