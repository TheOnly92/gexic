package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-gl/gl"
	"io/ioutil"
	"math"
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
	HEX_WIDTH  int = 92
	HEX_HEIGHT     = 80

	OddIncre = 39

	XMultiply = 67
	YMultiply = 77
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
	StateShrinking
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
			rt[x][y] = &Hex{HexType(rand.Intn(int(HexFlower)-1) + 1), StateNormal}
		}
	}
	for rt.CheckCollision() {
		for x := 0; x < 11; x++ {
			maxy := 8
			if x%2 == 1 {
				maxy = 9
			}
			for y := 0; y < maxy; y++ {
				if rt[x][y].State == StateShrinking {
					rt[x][y] = nil
					rt[x][y] = &Hex{HexType(rand.Intn(int(HexFlower)-1) + 1), StateNormal}
				}
			}
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
			m[x][y].Render(1, false)
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
	rt.X = float64(x * XMultiply)
	rt.Y = float64(y * YMultiply)
	if x%2 == 1 {
		rt.Y -= float64(OddIncre)
	}
	return rt
}

func (m HexMap) CheckCollision() bool {
	rt := false
	for x := 0; x < 11; x++ {
		maxy := 8
		if x%2 == 1 {
			maxy = 9
		}
		for y := 0; y < maxy; y++ {
			kind := m[x][y].Kind
			if y+1 < maxy && x < 10 {
				if x%2 == 0 && m[x+1][y+1].Kind == kind && m[x][y+1].Kind == kind {
					rt = true
					m[x][y].State = StateShrinking
					m[x+1][y+1].State = StateShrinking
					m[x][y+1].State = StateShrinking
				} else if x%2 == 1 && m[x+1][y].Kind == kind && m[x][y+1].Kind == kind {
					rt = true
					m[x][y].State = StateShrinking
					m[x+1][y].State = StateShrinking
					m[x][y+1].State = StateShrinking
				}
			}
			if x > 0 {
				fmt.Println(x, y)
				if x%2 == 0 && m[x-1][y+1].Kind == kind && m[x-1][y].Kind == kind {
					rt = true
					m[x][y].State = StateShrinking
					m[x-1][y+1].State = StateShrinking
					m[x-1][y].State = StateShrinking
				} else if x%2 == 1 && y > 0 && y < maxy-1 && m[x-1][y].Kind == kind && m[x-1][y-1].Kind == kind {
					rt = true
					m[x][y].State = StateShrinking
					m[x-1][y].State = StateShrinking
					m[x-1][y-1].State = StateShrinking
				}
			}
			if y > 0 && x < 10 {
				if x%2 == 0 && m[x+1][y].Kind == kind && m[x][y-1].Kind == kind {
					rt = true
					m[x][y].State = StateShrinking
					m[x+1][y].State = StateShrinking
					m[x][y-1].State = StateShrinking
				} else if x%2 == 1 && y > 0 && m[x+1][y-1].Kind == kind && m[x][y-1].Kind == kind {
					rt = true
					m[x][y].State = StateShrinking
					m[x+1][y-1].State = StateShrinking
					m[x][y-1].State = StateShrinking
				}
			}
			if y+1 < maxy && x > 0 {
				if x%2 == 0 && m[x-1][y+1].Kind == kind && m[x][y+1].Kind == kind {
					rt = true
					m[x][y].State = StateShrinking
					m[x-1][y+1].State = StateShrinking
					m[x][y+1].State = StateShrinking
				} else if x%2 == 1 && m[x-1][y].Kind == kind && m[x][y+1].Kind == kind {
					rt = true
					m[x][y].State = StateShrinking
					m[x-1][y].State = StateShrinking
					m[x][y+1].State = StateShrinking
				}
			}
			if x < 10 {
				if x%2 == 0 && m[x+1][y].Kind == kind && m[x+1][y+1].Kind == kind {
					rt = true
					m[x][y].State = StateShrinking
					m[x+1][y].State = StateShrinking
					m[x+1][y+1].State = StateShrinking
				} else if x%2 == 1 && y > 0 && y < maxy-1 && m[x+1][y-1].Kind == kind && m[x+1][y].Kind == kind {
					rt = true
					m[x][y].State = StateShrinking
					m[x+1][y-1].State = StateShrinking
					m[x+1][y].State = StateShrinking
				}
			}
			if y > 0 && x > 0 {
				if x%2 == 0 && m[x-1][y].Kind == kind && m[x][y-1].Kind == kind {
					rt = true
					m[x][y].State = StateShrinking
					m[x-1][y].State = StateShrinking
					m[x][y-1].State = StateShrinking
				} else if x%2 == 1 && m[x-1][y-1].Kind == kind && m[x][y-1].Kind == kind {
					rt = true
					m[x][y].State = StateShrinking
					m[x-1][y-1].State = StateShrinking
					m[x][y-1].State = StateShrinking
				}
			}
		}
	}
	return rt
}

func (m HexMap) CalcClosestCenter(x, y int) (Point, bool) {
	rt := Point{}
	hex := FieldPoint{int(math.Floor((float64(x) - FieldOffsetX) / float64(XMultiply))), 0}
	if hex.X%2 == 0 {
		hex.Y = int(math.Floor((float64(y) - FieldOffsetY) / float64(YMultiply)))
	} else {
		hex.Y = int(math.Floor((float64(y+OddIncre) - FieldOffsetY) / float64(YMultiply)))
	}
	left, right := m.GetCenter(0, 0), m.GetCenter(10, 0)
	x = int(math.Min(math.Max(float64(x), left.X), right.X))
	top, bottom := m.GetCenter(hex.X, 0), m.GetCenter(hex.X, 7)
	if hex.X%2 == 1 {
		bottom = m.GetCenter(hex.X, 8)
	}
	y = int(math.Min(math.Max(float64(y), top.Y), bottom.Y))
	if hex.X > 10 || hex.X < 0 || hex.Y < 0 || (hex.X%2 == 0 && hex.Y > 7) || (hex.X%2 == 1 && hex.Y > 8) {
		return rt, false
	}
	for loopX := hex.X - 1; loopX <= hex.X+2; loopX++ {
		if loopX < 0 || loopX > 10 {
			continue
		}
		for loopY := hex.Y - 2; loopY <= hex.Y; loopY++ {
			if loopY < 0 || (loopX%2 == 0 && loopY > 7) || (loopX%2 == 1 && loopY > 8) {
				continue
			}
			if loopX%2 == 1 {
				c1X, c1Y := m.GetCenter(loopX, loopY).WithOffset()
				c2X, c2Y := m.GetCenter(loopX+1, loopY).WithOffset()
				c3X, c3Y := m.GetCenter(loopX, loopY+1).WithOffset()
				c4X, c4Y := m.GetCenter(loopX+1, loopY-2).WithOffset()
				if pointInTriangle(float64(x), float64(y), c1X, c1Y, c2X, c2Y, c3X, c3Y) {
					mouse.selectedHex = nil
					mouse.selectedHex = []FieldPoint{
						{loopX, loopY},
						{loopX + 1, loopY},
						{loopX, loopY + 1},
					}
					rt.X = c2X - float64(HEX_WIDTH)/2
					rt.Y = c2Y
					return rt, true
				} else if pointInTriangle(float64(x), float64(y), c1X, c1Y, c4X, c4Y, c2X, c2Y) && loopX+1 < 11 && loopY-1 >= 0 {
					mouse.selectedHex = nil
					mouse.selectedHex = []FieldPoint{
						{loopX, loopY},
						{loopX + 1, loopY - 1},
						{loopX + 1, loopY},
					}
					rt.X = c1X - float64(HEX_WIDTH)/2
					rt.Y = c1Y
					return rt, true
				}
			} else {
				c1X, c1Y := m.GetCenter(loopX, loopY).WithOffset()
				c2X, c2Y := m.GetCenter(loopX+1, loopY+1).WithOffset()
				c3X, c3Y := m.GetCenter(loopX, loopY+1).WithOffset()
				c4X, c4Y := m.GetCenter(loopX+1, loopY).WithOffset()
				if pointInTriangle(float64(x), float64(y), c1X, c1Y, c2X, c2Y, c3X, c3Y) {
					mouse.selectedHex = nil
					mouse.selectedHex = []FieldPoint{
						{loopX, loopY},
						{loopX + 1, loopY + 1},
						{loopX, loopY + 1},
					}
					rt.X = c2X - float64(HEX_WIDTH)/2
					rt.Y = c2Y
					return rt, true
				} else if pointInTriangle(float64(x), float64(y), c1X, c1Y, c4X, c4Y, c2X, c2Y) {
					mouse.selectedHex = nil
					mouse.selectedHex = []FieldPoint{
						{loopX, loopY},
						{loopX + 1, loopY},
						{loopX + 1, loopY + 1},
					}
					rt.X = c1X - float64(HEX_WIDTH)/2
					rt.Y = c1Y
					return rt, true
				}
			}
		}
	}
	return rt, false
}

func (h *Hex) Render(alpha float32, drawFromCenter bool) {
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
	if drawFromCenter {
		gl.Vertex2i(HEX_WIDTH/2, HEX_HEIGHT/2)
	} else {
		gl.Vertex2i(HEX_WIDTH, HEX_HEIGHT)
	}
	gl.TexCoord2f(0, 1)
	if drawFromCenter {
		gl.Vertex2i(HEX_WIDTH/2, -HEX_HEIGHT/2)
	} else {
		gl.Vertex2i(HEX_WIDTH, 0)
	}
	gl.TexCoord2f(1, 1)
	if drawFromCenter {
		gl.Vertex2i(-HEX_WIDTH/2, -HEX_HEIGHT/2)
	} else {
		gl.Vertex2i(0, 0)
	}
	gl.TexCoord2f(1, 0)
	if drawFromCenter {
		gl.Vertex2i(-HEX_WIDTH/2, HEX_HEIGHT/2)
	} else {
		gl.Vertex2i(0, HEX_HEIGHT)
	}
	gl.End()
}
