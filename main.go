package main

import (
	"fmt"
	"github.com/go-gl/gl"
	"github.com/go-gl/glfw"
	// "github.com/go-gl/glu"
	"image"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"reflect"
	"time"
	// "reflect"
)

var hexTex, wallpaper, starTex, borderTex gl.Texture

var hexMap2 HexMap
var hexMap [10][9]int
var currExX, currExY int
var rotate float32
var timesToRotate int
var currentMatches [][]int
var starMatches [][]int
var starScale float32
var starAlpha float32
var starRotate bool
var currStarCenter []int
var scale float32
var animateFall []*freefall
var fallticks int
var mouseLock bool
var mousePos []int
var selectedHex [][]int
var posX, posY int
var prevSelectPos []int

type freefall struct {
	x, y    int
	targetY int
	accel   float64
}

type Point struct {
	X, Y float64
}

func (p Point) WithOffset() (float64, float64) {
	return p.X + 80, p.Y + 80
}

type FieldPoint struct {
	X, Y int
}

func genHexMap() {
	// hexMap = [10][9]int{
	// 	[9]int{0, 2, 0, 0, 5, 1, 1, 1, -1},
	// 	[9]int{1, 1, 3, 5, 2, 1, 4, 3, 2},
	// 	[9]int{5, 4, 2, 0, 3, 1, 1, 0, -1},
	// 	[9]int{3, 4, 1, 3, 2, 5, 2, 3, 4},
	// 	[9]int{4, 3, 3, 5, 3, 4, 1, 5, -1},
	// 	[9]int{1, 3, 2, 1, 2, 3, 1, 4, 1},
	// 	[9]int{4, 5, 6, 5, 1, 5, 3, 6, -1},
	// 	[9]int{0, 5, 4, 3, 4, 3, 0, 2, 3},
	// 	[9]int{3, 2, 4, 5, 2, 5, 0, 4, -1},
	// 	[9]int{0, 2, 5, 0, 0, 2, 2, 4, 5}}
	// return
	rand.Seed(time.Now().Unix())
	for x := 0; x < 10; x++ {
		maxy := 8
		hexMap[x][8] = -1
		if x%2 == 1 {
			maxy = 9
		}
		for y := 0; y < maxy; y++ {
			hexMap[x][y] = rand.Intn(6)
			// hexMap[x][y] = 6
		}
	}
}

func removeHexAndGenNew(matched [][]int) {
	// fmt.Println("removeHexAndGenNew ", matched)
	for _, v := range matched {
		hexMap[v[0]][v[1]] = -1
	}
	for x := 0; x < 10; x++ {
		maxy := 7
		if x%2 == 1 {
			maxy = 8
		}
		for y := maxy; y >= 0; y-- {
			if hexMap[x][y] == -1 {
				for y2 := y; y2 > 0; y2-- {
					hexMap[x][y2] = hexMap[x][y2-1]
				}
				hexMap[x][0] = rand.Intn(6)
				y = maxy + 1
			}
		}
	}
}

func getStarCenter(matched [][]int) []int {
	x := 0
	miny := 99
	fmt.Println(matched)
	for _, v := range matched {
		// hexMap[v[0]][v[1]] = -1
		x += v[0]
		if v[1] < miny {
			miny = v[1]
		}
	}
	if miny < 99 {
		return []int{x / 6, miny + 1}
	}
	return []int{-1, -1}
}

func makeStarAndGenNew(matched [][]int) {
	center := [][]int{}
	x := 0
	miny := 99
	fmt.Println(matched)
	for i, v := range matched {
		if i%6 == 0 && i > 0 {
			center = append(center, []int{x / 6, miny + 1})
			x = 0
			miny = 99
		}
		hexMap[v[0]][v[1]] = -1
		x += v[0]
		if v[1] < miny {
			miny = v[1]
		}
	}
	if miny < 99 {
		center = append(center, []int{x / 6, miny + 1})
	}
	for _, v := range center {
		fmt.Println(v)
		hexMap[v[0]][v[1]] = 6
	}
	for x := 0; x < 10; x++ {
		maxy := 7
		if x%2 == 1 {
			maxy = 8
		}
		for y := maxy; y >= 0; y-- {
			if hexMap[x][y] == -1 {
				for y2 := y; y2 > 0; y2-- {
					hexMap[x][y2] = hexMap[x][y2-1]
				}
				hexMap[x][0] = rand.Intn(6)
				y = maxy + 1
			}
		}
	}
}

func getFallTargetY(x, y int) int {
	maxy := 7
	if x%2 == 1 {
		maxy = 8
	}
	add := 0
	for yi := maxy; yi > y; yi-- {
		found := false
		for _, v := range append(currentMatches, starMatches...) {
			if v[0] == x && v[1] == yi {
				found = true
				break
			}
		}
		if found {
			add++
		}
	}
	return y + add
}

func renderHexMap() {
	gl.Enable(gl.TEXTURE_2D)
	gl.Enable(gl.BLEND)
	gl.Disable(gl.DEPTH_TEST)
	wallpaper.Bind(gl.TEXTURE_2D)
	gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.REPLACE)
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 0)
	gl.Vertex2i(0, 0)
	gl.TexCoord2f(0, 1)
	gl.Vertex2i(0, 768)
	gl.TexCoord2f(1, 1)
	gl.Vertex2i(1024, 768)
	gl.TexCoord2f(1, 0)
	gl.Vertex2i(1024, 0)
	gl.End()
	gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.DECAL)
	gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.PushMatrix()
	gl.Translatef(80, 80, 0)
	for x := 0; x < 10; x++ {
		maxy := 8
		topy := 19
		if x%2 == 1 {
			maxy = 9
			topy = 0
		}
		starty := 0
		for y := starty; y < maxy; y++ {
			if currExX > -1 && currExY > -1 && starRotate {
				if y == currExY && x == currExX || y == currExY+1 && x == currExX || y == currExY-1 && x == currExX || currExX%2 == 0 && ((currExX == x-1 || currExX == x+1) && currExY == y-1 || (currExX == x-1 || currExX == x+1) && currExY == y) || currExX%2 == 1 && ((currExX == x-1 || currExX == x+1) && currExY == y || (currExX == x-1 || currExX == x+1) && currExY == y+1) {
					continue
				}
			} else if timesToRotate > 0 {
				// if y == currExY && x == currExX || currExX%2 == 0 && (x == currExX+1 && y == currExY || x == currExX+1 && y == currExY+1) || currExX%2 == 1 && (x == currExX+1 && y == currExY || x == currExX+1 && y == currExY-1) {
				// 	continue
				// }
				found := false
				for _, v := range selectedHex {
					if y == v[1] && x == v[0] {
						found = true
						break
					}
				}
				if found {
					continue
				}
			}
			found := false
			for _, v := range currentMatches {
				if scale > 0 && v[0] == x && v[1] == y || scale <= 0 && v[0] == x && v[1] >= y {
					found = true
					break
				}
			}
			for _, v := range starMatches {
				if starAlpha > 0 && v[0] == x && v[1] == y || starAlpha <= 0 && v[0] == x && v[1] >= y {
					found = true
					break
				}
			}
			if found || len(currStarCenter) > 0 && currStarCenter[0] == x && currStarCenter[1] == y {
				continue
			}
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			drawHex(x*33, topy+y*38, hexMap[x][y], 1)
		}
	}
	gl.PopMatrix()
	if len(currentMatches) > 0 || len(starMatches) > 0 {
		mouseLock = true
		if len(currentMatches) > 0 && scale > 0 {
			scale -= 0.1
			for _, v := range currentMatches {
				gl.PushMatrix()
				topy := 19
				if v[0]%2 == 1 {
					topy = 0
				}
				gl.Translatef(float32(v[0]*33+102), float32(v[1]*38+topy+94), 0)
				gl.Scalef(scale, scale, 1)
				gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
				drawHex(-22, -14, hexMap[v[0]][v[1]], 1)
				gl.PopMatrix()
			}
		} else if len(starMatches) > 0 && starAlpha > 0 {
			starAlpha -= 0.05
			// starAlpha = 0.7
			starScale += 0.05
			// starScale = 1.4
			gl.PushMatrix()
			topy := 19
			pm := 0
			if currStarCenter[0]%2 == 1 {
				topy = 0
				pm = -1
			}
			gl.Translatef(float32(currStarCenter[0]*33+102), float32(currStarCenter[1]*38+topy+94), 0)
			drawHex(-22, -14, 6, 1)
			gl.PopMatrix()
			gl.PushMatrix()
			gl.Translatef(float32(currStarCenter[0]*33+102), float32(currStarCenter[1]*38+topy+94), 0)
			gl.Scalef(starScale, starScale, 1)
			drawHex(-22, -14-HEX_HEIGHT, hexMap[currStarCenter[0]][currStarCenter[1]-1], starAlpha)
			drawHex(-22, -20+HEX_HEIGHT, hexMap[currStarCenter[0]][currStarCenter[1]+1], starAlpha)
			drawHex(-52, -36, hexMap[currStarCenter[0]-1][currStarCenter[1]+pm], starAlpha)
			drawHex(-52, -40+HEX_HEIGHT, hexMap[currStarCenter[0]-1][currStarCenter[1]+pm+1], starAlpha)
			drawHex(8, -36, hexMap[currStarCenter[0]+1][currStarCenter[1]+pm], starAlpha)
			drawHex(8, -40+HEX_HEIGHT, hexMap[currStarCenter[0]+1][currStarCenter[1]+pm+1], starAlpha)
			gl.PopMatrix()
		} else {
			if fallticks == 0 {
				animateFall = make([]*freefall, 0)
				for x := 0; x < 10; x++ {
					topy := 19
					if x%2 == 1 {
						topy = 0
					}
					fromy := -1
					for _, v := range currentMatches {
						if v[0] != x {
							continue
						}
						if v[1] > fromy {
							fromy = v[1]
						}
					}
					for _, v := range starMatches {
						if v[0] != x {
							continue
						}
						if v[1] > fromy {
							fromy = v[1]
						}
					}
					if fromy == -1 {
						continue
					}
					for y := fromy; y >= 0; y-- {
						found := false
						for _, v := range currentMatches {
							if v[0] == x && v[1] == y {
								found = true
								break
							}
						}
						for _, v := range starMatches {
							if v[0] == x && v[1] == y {
								found = true
								break
							}
						}
						if found {
							continue
						}
						animateFall = append(animateFall, &freefall{x, y, getFallTargetY(x, y), math.Pow(float64(y), 2)/16 + 0.5})
						gl.PushMatrix()
						gl.Translatef(float32(x*33+102), float32(y*38+topy+94), 0)
						gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
						drawHex(-22, -14, hexMap[x][y], 1)
						gl.PopMatrix()
					}
				}
				fallticks++
			} else {
				stillFalling := 0
				for _, v := range animateFall {
					topy := 19
					if v.x%2 == 1 {
						topy = 0
					}
					newy := v.accel * math.Pow(float64(fallticks), 2) / 2
					gl.PushMatrix()
					gl.Translatef(float32(v.x*33+102), float32(math.Min(float64(v.y*38+topy+94)+newy, float64(v.targetY*38+topy+94))), 0)
					gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
					drawHex(-22, -14, hexMap[v.x][v.y], 1)
					gl.PopMatrix()
					if float64(v.y*38+topy+94)+newy < float64(v.targetY*38+topy+94) {
						stillFalling++
					}
				}
				fallticks++
				if stillFalling == 0 {
					starScale = 1
					starAlpha = 0
					removeHexAndGenNew(currentMatches)
					makeStarAndGenNew(starMatches)
					currentMatches = checkHexMap()
					starMatches = checkHexStar()
					currStarCenter = []int{}
					scale = 1
					fallticks = 0
					mouseLock = false
					fmt.Println("Mouse unlocked 1")
					animateFall = nil
				}
			}
		}
	}
	if currExX > -1 && currExY > -1 {
		gl.PushMatrix()
		topy := 19
		if currExX%2 == 1 {
			topy = 0
		}
		if starRotate {
			timesToRotate = 0
			gl.Translatef(float32(currExX*33+102), float32(currExY*38+topy+94), 0)
			gl.Scalef(1.3, 1.3, 1)
			gl.Rotatef(rotate, 0, 0, 1)
			drawHex(-22, -14, 6, 1)
			drawHex(-22, -14-HEX_HEIGHT, hexMap[currExX][currExY-1], 1)
			drawHex(-22, -20+HEX_HEIGHT, hexMap[currExX][currExY+1], 1)
			pm := 0
			if currExX%2 == 1 {
				pm = -1
			}
			drawHex(-52, -36, hexMap[currExX-1][currExY+pm], 1)
			drawHex(-52, -40+HEX_HEIGHT, hexMap[currExX-1][currExY+pm+1], 1)
			drawHex(8, -36, hexMap[currExX+1][currExY+pm], 1)
			drawHex(8, -40+HEX_HEIGHT, hexMap[currExX+1][currExY+pm+1], 1)
		} else {
			// gl.Translatef(float32(currExX*33+HEX_WIDTH+80), float32(currExxY*38+topy+20+80), 0)
			gl.Translatef(float32(prevSelectPos[0]), float32(prevSelectPos[1]), 0)
			gl.Scalef(1.3, 1.3, 1)
			gl.Rotatef(rotate, 0, 0, 1)
			for _, v := range selectedHex {
				switch v[2] {
				case 1:
					drawHex(-32, -34, hexMap[v[0]][v[1]], 1)
				case 2:
					drawHex(0, -17, hexMap[v[0]][v[1]], 1)
				case 3:
					drawHex(-32, 0, hexMap[v[0]][v[1]], 1)
				case 4:
					drawHex(-44, -19, hexMap[v[0]][v[1]], 1)
				case 5:
					drawHex(-12, -36, hexMap[v[0]][v[1]], 1)
				case 6:
					drawHex(-12, -2, hexMap[v[0]][v[1]], 1)
				}
			}
			// if currExX%2 == 0 {
			// 	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			// 	drawHex(-12, -36, hexMap[currExX+1][currExY], 1)
			// } else {
			// 	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			// 	drawHex(-12, -36, hexMap[currExX+1][currExY-1], 1)
			// }
			// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			// drawHex(-44, -19, hexMap[currExX][currExY], 1)
			// if currExX%2 == 0 {
			// 	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			// 	drawHex(-12, -2, hexMap[currExX+1][currExY+1], 1)
			// } else {
			// 	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			// 	drawHex(-12, -2, hexMap[currExX+1][currExY], 1)
			// }
		}
		gl.PopMatrix()
		if !starRotate && rotate < 120 {
			rotate += 15
			// rotate = 15
		} else if starRotate && rotate < 60 {
			rotate += 6
		} else {
			if starRotate {
				if currExX%2 == 0 {
					hexMap[currExX][currExY-1], hexMap[currExX+1][currExY], hexMap[currExX+1][currExY+1], hexMap[currExX][currExY+1], hexMap[currExX-1][currExY+1], hexMap[currExX-1][currExY] = hexMap[currExX-1][currExY], hexMap[currExX][currExY-1], hexMap[currExX+1][currExY], hexMap[currExX+1][currExY+1], hexMap[currExX][currExY+1], hexMap[currExX-1][currExY+1]
				} else {
					hexMap[currExX][currExY-1], hexMap[currExX+1][currExY-1], hexMap[currExX+1][currExY], hexMap[currExX][currExY+1], hexMap[currExX-1][currExY], hexMap[currExX-1][currExY-1] = hexMap[currExX-1][currExY-1], hexMap[currExX][currExY-1], hexMap[currExX+1][currExY-1], hexMap[currExX+1][currExY], hexMap[currExX][currExY+1], hexMap[currExX-1][currExY]
				}
			} else {
				v2 := make([][]int, 3)
				for _, v := range selectedHex {
					idx := 0
					switch v[2] {
					case 1, 4:
						idx = 0
					case 2, 5:
						idx = 1
					case 3, 6:
						idx = 2
					}
					v2[idx] = []int{v[0], v[1]}
					fmt.Println(idx, v[0], v[1])
				}
				hexMap[v2[0][0]][v2[0][1]], hexMap[v2[1][0]][v2[1][1]], hexMap[v2[2][0]][v2[2][1]] = hexMap[v2[2][0]][v2[2][1]], hexMap[v2[0][0]][v2[0][1]], hexMap[v2[1][0]][v2[1][1]]
			}
			starMatches = checkHexStar()
			if len(starMatches) > 6 {
				fmt.Println(starMatches)
			}
			if len(starMatches) >= 6 {
				timesToRotate = 0
				rotate = 0
				currExX = -1
				currExY = -1
				starScale = 1
				starAlpha = 1
				starRotate = false
				currStarCenter = getStarCenter(starMatches)
				hexMap[currStarCenter[0]][currStarCenter[1]] = 6
				// makeStarAndGenNew(starMatches)
			} else {
				matches := checkHexMap()
				if len(matches) > 0 {
					currentMatches = matches
					scale = 1
					currExX = -1
					currExY = -1
					rotate = 0
					timesToRotate = 0
					starRotate = false
				} else {
					if timesToRotate == 0 {
						currExX = -1
						currExY = -1
						rotate = 0
						timesToRotate = 0
						starRotate = false
						mouseLock = false
						fmt.Println("Mouse unlocked 3")
					}
					rotate = 0
					timesToRotate--
				}
			}
		}
	}
	if !mouseLock {
		prevSelectPos = calcClosestCenter(posX, posY)
		drawBorderAtXY(float32(prevSelectPos[0]), float32(prevSelectPos[1]), prevSelectPos[2])
	}
	gl.Flush()
	gl.Disable(gl.TEXTURE_2D)
	gl.Disable(gl.BLEND)
}

func pointInTriangle(pX, pY, aX, aY, bX, bY, cX, cY float64) bool {
	bc := bX*cY - bY*cX
	ca := cX*aY - cY*aX
	ab := aX*bY - aY*bX
	ap := aX*pY - aY*pX
	bp := bX*pY - bY*pX
	cp := cX*pY - cY*pX
	abc := float64(0)
	if bc+ca+ab > 0 {
		abc = 1
	} else if bc+ca+ab < 0 {
		abc = -1
	}
	if abc*(bc-bp+cp) > 0 {
		if abc*(ca-cp+ap) > 0 {
			if abc*(ab-ap+bp) > 0 {
				return true
			}
		}
	}
	return false
}

func hexCenter(x, y int) []int {
	rt := []int{0, 0}
	rt[0] = x*33 + 80 + HEX_WIDTH/2
	topy := 13
	if x%2 == 1 {
		topy = -8
	}
	rt[1] = y*38 + 80 + topy + HEX_HEIGHT/2
	return rt
}

func calcClosestCenter(x, y int) []int {
	hexX := int(math.Floor((float64(x) - 80) / 30))
	topy := -19
	if hexX%2 == 1 {
		topy = 0
	}
	left := hexCenter(0, 0)
	right := hexCenter(9, 0)
	x = int(math.Min(math.Max(float64(x), float64(left[0])), float64(right[0])))
	top := hexCenter(hexX, 0)
	bottom := hexCenter(hexX, 7)
	if hexX%2 == 1 {
		bottom = hexCenter(hexX, 8)
	}
	y = int(math.Min(math.Max(float64(y), float64(top[1])), float64(bottom[1])))
	if hexX%2 == 1 && y > bottom[1]-10 {
		x = bottom[0]
	}
	hexY := int(math.Floor((float64(y+topy) - 80) / 36))
	if hexX > 9 || hexY > 8 || hexX < 0 || hexY < 0 {
		return []int{-1, -1, -1}
	}
	rt := prevSelectPos
	for loopX := hexX - 1; loopX <= hexX+2; loopX++ {
		if loopX >= 0 && loopX < 10 {
			for loopY := hexY - 2; loopY <= hexY; loopY++ {
				if loopY >= 0 && (loopX%2 == 0 && loopY < 8 || loopX%2 == 1 && loopY < 9) {
					if loopX%2 == 1 {
						c1 := hexCenter(loopX, loopY)
						c2 := hexCenter(loopX+1, loopY)
						c3 := hexCenter(loopX, loopY+1)
						c4 := hexCenter(loopX+1, loopY-1)
						if pointInTriangle(float64(x), float64(y), float64(c1[0]), float64(c1[1]), float64(c2[0]), float64(c2[1]), float64(c3[0]), float64(c3[1])) {
							selectedHex = [][]int{[]int{loopX, loopY, 1}, []int{loopX + 1, loopY, 2}, []int{loopX, loopY + 1, 3}}
							// fmt.Println(1, c2[1], loopX+1)
							return []int{c2[0] - HEX_WIDTH/2, c2[1], 1}
						} else if pointInTriangle(float64(x), float64(y), float64(c1[0]), float64(c1[1]), float64(c4[0]), float64(c4[1]), float64(c2[0]), float64(c2[1])) && loopX+1 < 10 && loopY-1 >= 0 {
							selectedHex = [][]int{[]int{loopX, loopY, 4}, []int{loopX + 1, loopY - 1, 5}, []int{loopX + 1, loopY, 6}}
							// fmt.Println(2, c1[1], loopX)
							return []int{c1[0] + HEX_WIDTH/2 - 7, c1[1] + 4, 0}
						}
					} else {
						c1 := hexCenter(loopX, loopY)
						c2 := hexCenter(loopX+1, loopY+1)
						c3 := hexCenter(loopX, loopY+1)
						c4 := hexCenter(loopX+1, loopY)
						if pointInTriangle(float64(x), float64(y), float64(c1[0]), float64(c1[1]), float64(c2[0]), float64(c2[1]), float64(c3[0]), float64(c3[1])) {
							selectedHex = [][]int{[]int{loopX, loopY, 1}, []int{loopX + 1, loopY + 1, 2}, []int{loopX, loopY + 1, 3}}
							// fmt.Println(3, c2[1], loopX+1)
							return []int{c2[0] - HEX_WIDTH/2, c2[1], 1}
						} else if pointInTriangle(float64(x), float64(y), float64(c1[0]), float64(c1[1]), float64(c4[0]), float64(c4[1]), float64(c2[0]), float64(c2[1])) {
							selectedHex = [][]int{[]int{loopX, loopY, 4}, []int{loopX + 1, loopY, 5}, []int{loopX + 1, loopY + 1, 6}}
							// fmt.Println(4, c1[1], loopX)
							return []int{c1[0] + HEX_WIDTH/2 - 7, c1[1] + 1, 0}
						}
					}
				}
			}
		}
	}
	return rt
}

func drawBorderAtXY(x, y float32, reverse int) {
	if x <= 80 || y <= 80 {
		return
	}
	gl.PushMatrix()
	gl.Translatef(x, y, 0)
	if reverse == 1 {
		gl.Rotatef(60, 0, 0, 1)
	}
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.REPLACE)
	borderTex.Bind(gl.TEXTURE_2D)
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 0)
	gl.Vertex2i(-38, -38)
	gl.TexCoord2f(0, 1)
	gl.Vertex2i(-38, 38)
	gl.TexCoord2f(1, 1)
	gl.Vertex2i(38, 38)
	gl.TexCoord2f(1, 0)
	gl.Vertex2i(38, -38)
	gl.End()
	gl.PopMatrix()
}

func main() {
	sys := Make()
	sys.Startup()
	defer sys.Shutdown()
	// InitQueue()

	sys.CreateWindow(1024, 768, "Gexic")
	gl.ClearColor(0., 0.2, 0.4, 0.)
	initGL()

	prevSelectPos = []int{0, 0, 0}

	// PurgeQueue()
	genHexMap()
	hexMap2 = GenHexMap()
	for matches := checkHexMap(); len(matches) > 0; matches = checkHexMap() {
		removeHexAndGenNew(matches)
	}
	glfw.SetMouseButtonCallback(mouseButtonCallback)
	glfw.SetCharCallback(charCallback)
	glfw.SetMousePosCallback(mousePosCallback)
	currExX = -1
	currExY = -1

	for sys.CheckExitMainLoop() {
		start := glfw.Time()
		wait := float64(1) / float64(30)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		// renderHexMap()
		hexMap2.Render()
		sys.Refresh()
		diff := glfw.Time() - start
		if diff < wait {
			glfw.Sleep(wait - diff)
		}
	}
}

func initGL() {
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, 1024, 768, 0, -1, 1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Enable(gl.DEPTH_TEST)

	// hexTex = initTexture("hex5k", HEX_WIDTH, HEX_HEIGHT)
	hexTex = initTexture2("hex7d")
	wallpaper = initTexture("wallpaper-2594238", 1024, 768)
	// starTex = initTexture("hexstark", HEX_WIDTH, HEX_HEIGHT)
	starTex = initTexture2("hex0")
	borderTex = initTexture("hexborder", 76, 76)
}

func initTexture2(filename string) gl.Texture {
	img, err := glfw.ReadImage(filename+".tga", 0)
	if err != nil {
		panic(err)
	}
	rt := gl.GenTexture()
	gl.Enable(gl.TEXTURE_2D)
	rt.Bind(gl.TEXTURE_2D)
	gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	// gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	// gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, img.Width(), img.Height(), 0, gl.RGBA, gl.UNSIGNED_BYTE, img.Data())
	fmt.Println(img.Width(), img.Height())
	return rt
}

func initTexture(filename string, width, height int) gl.Texture {
	file, err := os.Open(filename + ".png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}
	t := reflect.ValueOf(img)
	fmt.Println(t.Elem().Type().Name())
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if (filename == "hex4v" || filename == "hexstar2" || filename == "hexborder") && r == 0 && g == 0 && b == 0 {
				a = 0
			}
			// if filename == "hex5k" {
			// 	fmt.Println(r, g, b, a)
			// }
			base := 4*x + canvas.Stride*y
			canvas.Pix[base] = uint8(r)
			canvas.Pix[base+1] = uint8(g)
			canvas.Pix[base+2] = uint8(b)
			canvas.Pix[base+3] = uint8(a)
		}
	}
	rt := gl.GenTexture()
	gl.Enable(gl.TEXTURE_2D)
	rt.Bind(gl.TEXTURE_2D)
	gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	// gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	// gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, width, height, 0, gl.RGBA, gl.UNSIGNED_BYTE, canvas.Pix)
	return rt
}

func drawHex(x, y, kind int, alpha float32) {
	if kind == 6 {
		gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.REPLACE)
		starTex.Bind(gl.TEXTURE_2D)
		gl.Begin(gl.QUADS)
		gl.TexCoord2f(0, 0)
		gl.Vertex2i(x, y)
		gl.TexCoord2f(0, 1)
		gl.Vertex2i(x, y+HEX_HEIGHT)
		gl.TexCoord2f(1, 1)
		gl.Vertex2i(x+HEX_WIDTH, y+HEX_HEIGHT)
		gl.TexCoord2f(1, 0)
		gl.Vertex2i(x+HEX_WIDTH, y)
		gl.End()
	} else {
		gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		var r, g, b float32
		switch kind {
		case 0:
			r = 1
		case 1:
			g = 1
		case 2:
			b = 1
		case 3:
			r = 1
			g = 1
		case 4:
			r = 1
			b = 1
		case 5:
			g = 1 - 222/255
			b = 1
		}
		hexTex.Bind(gl.TEXTURE_2D)
		gl.Begin(gl.QUADS)
		if alpha < 1 {
			gl.Color4f(r, g, b, alpha)
		} else {
			gl.Color3f(r, g, b)
		}
		gl.TexCoord2f(0, 0)
		gl.Vertex2i(x, y)
		gl.TexCoord2f(0, 1)
		gl.Vertex2i(x, y+HEX_HEIGHT)
		gl.TexCoord2f(1, 1)
		gl.Vertex2i(x+HEX_WIDTH, y+HEX_HEIGHT)
		gl.TexCoord2f(1, 0)
		gl.Vertex2i(x+HEX_WIDTH, y)
		gl.End()
	}
}

func mousePosCallback(x, y int) {
	if mouseLock {
		return
	}
	posX = x
	posY = y
}

func mouseButtonCallback(button, state int) {
	if currExX != -1 || currExY != -1 || mouseLock || len(selectedHex) < 3 {
		return
	}
	x, y := glfw.MousePos()

	if state == glfw.KeyPress {
		switch button {
		case glfw.MouseLeft:
			// fmt.Println(x, y)
			currExX = int(math.Floor((float64(x) - 80) / 33))
			currExY = int(math.Floor((float64(y) - 80 - 19) / 38))
			if currExX%2 == 1 {
				currExY = (y - 80) / 36
			}
			if currExX > 9 || currExY > 8 || currExX < 0 || currExY < 0 {
				currExX = -1
				currExY = -1
				return
			}
			if hexMap[currExX][currExY] == 6 && currExX > 0 && currExX < 9 && currExY > 0 && (currExX%2 == 0 && currExY < 7 || currExX%2 == 1 && currExY < 8) {
				starRotate = true
			}
			timesToRotate = 2
			mouseLock = true
			fmt.Println("Mouse locked")
			// fmt.Println(currExX, currExY)
			// renderHexMap(currExX, currExY)
		}
	}
}

func charCallback(button, state int) {
	if state == glfw.KeyPress {
		if button == 'a' {
			fmt.Println(hexMap)
		} else if button == 'r' {
			genHexMap()
		}
	}
}

func checkHexStar() [][]int {
	var matched [][]int
	for x := 0; x < 10; x++ {
		maxy := 8
		if x%2 == 1 {
			maxy = 9
		}
		for y := 0; y < maxy; y++ {
			kind := hexMap[x][y]
			if y+2 < maxy && x < 9 && x > 0 {
				if x%2 == 0 {
					if hexMap[x-1][y+1] == kind && hexMap[x+1][y+1] == kind && hexMap[x-1][y+2] == kind && hexMap[x+1][y+2] == kind && hexMap[x][y+2] == kind {
						matched = append(matched, []int{x, y}, []int{x - 1, y + 1}, []int{x + 1, y + 1}, []int{x - 1, y + 2}, []int{x + 1, y + 2}, []int{x, y + 2})
					}
				} else {
					if hexMap[x-1][y] == kind && hexMap[x+1][y] == kind && hexMap[x-1][y+1] == kind && hexMap[x+1][y+1] == kind && hexMap[x][y+2] == kind {
						matched = append(matched, []int{x, y}, []int{x - 1, y}, []int{x + 1, y}, []int{x - 1, y + 1}, []int{x + 1, y + 1}, []int{x, y + 2})
					}
				}
			}
			if x-2 >= 0 {
				if x%2 == 0 {
					if y+2 < maxy {
						if hexMap[x-1][y] == kind && hexMap[x-2][y] == kind && hexMap[x-2][y+1] == kind && hexMap[x][y+1] == kind && hexMap[x-1][y+2] == kind {
							matched = append(matched, []int{x, y}, []int{x - 1, y}, []int{x - 2, y}, []int{x - 2, y + 1}, []int{x, y + 1}, []int{x - 1, y + 2})
						}
					}
					if y+1 < maxy && y-1 >= 0 {
						if hexMap[x-1][y+1] == kind && hexMap[x-2][y] == kind && hexMap[x-2][y-1] == kind && hexMap[x-1][y-1] == kind && hexMap[x][y-1] == kind {
							matched = append(matched, []int{x, y}, []int{x - 1, y + 1}, []int{x - 2, y}, []int{x - 2, y - 1}, []int{x - 1, y - 1}, []int{x, y - 1})
						}
					}
				} else if y+1 < maxy {
					if y-1 >= 0 && y+1 < maxy {
						if hexMap[x-1][y-1] == kind && hexMap[x-2][y] == kind && hexMap[x][y+1] == kind && hexMap[x-2][y+1] == kind && hexMap[x-1][y+1] == kind {
							matched = append(matched, []int{x, y}, []int{x - 1, y - 1}, []int{x - 2, y}, []int{x, y + 1}, []int{x - 2, y + 1}, []int{x - 1, y + 1})
						}
					}
					if y-2 >= 0 {
						if hexMap[x-1][y] == kind && hexMap[x-2][y] == kind && hexMap[x-2][y-1] == kind && hexMap[x-1][y-2] == kind && hexMap[x][y-1] == kind {
							matched = append(matched, []int{x, y}, []int{x - 1, y}, []int{x - 2, y}, []int{x - 2, y - 1}, []int{x - 1, y - 2}, []int{x, y - 1})
						}
					}
				}
			}
		}
	}
	if len(matched) > 0 {
		fmt.Println(matched)
	}
	var rt [][]int
	for _, v := range matched {
		found := false
		for _, v2 := range rt {
			if v2[0] == v[0] && v2[1] == v[1] {
				found = true
				break
			}
		}
		if !found {
			rt = append(rt, v)
		}
	}
	if len(rt) != 6 && len(rt) > 0 {
		fmt.Println("After remove duplicate,", rt)
		return [][]int{}
	}
	return rt
}

func checkHexMap() [][]int {
	// return [][]int{}
	var matched [][]int
	for x := 0; x < 10; x++ {
		maxy := 8
		if x%2 == 1 {
			maxy = 9
		}
		for y := 0; y < maxy; y++ {
			kind := hexMap[x][y]
			// 1
			//     2
			// 3
			// Check 1
			if y+1 < maxy && x < 9 {
				if x%2 == 0 {
					if hexMap[x+1][y+1] == kind && hexMap[x][y+1] == kind {
						matched = append(matched, []int{x, y}, []int{x + 1, y + 1}, []int{x, y + 1})
					}
				} else {
					if hexMap[x+1][y] == kind && hexMap[x][y+1] == kind {
						matched = append(matched, []int{x, y}, []int{x + 1, y}, []int{x, y + 1})
					}
				}
			}
			// Check 2
			if x > 0 {
				if x%2 == 0 {
					if hexMap[x-1][y+1] == kind && hexMap[x-1][y] == kind {
						matched = append(matched, []int{x, y}, []int{x - 1, y + 1}, []int{x - 1, y})
					}
				} else if y > 0 {
					if hexMap[x-1][y] == kind && hexMap[x-1][y-1] == kind {
						matched = append(matched, []int{x, y}, []int{x - 1, y}, []int{x - 1, y - 1})
					}
				}
			}
			// Check 3
			if y > 0 && x < 9 {
				if x%2 == 0 {
					if hexMap[x+1][y] == kind && hexMap[x][y-1] == kind {
						matched = append(matched, []int{x, y}, []int{x + 1, y}, []int{x, y - 1})
					}
				} else if y > 0 {
					if hexMap[x+1][y-1] == kind && hexMap[x][y-1] == kind {
						matched = append(matched, []int{x, y}, []int{x + 1, y - 1}, []int{x, y - 1})
					}
				}
			}
			//     4
			// 5
			//     6
			// Check 4
			if y+1 < maxy && x > 0 {
				if x%2 == 0 {
					if hexMap[x-1][y+1] == kind && hexMap[x][y+1] == kind {
						matched = append(matched, []int{x, y}, []int{x - 1, y + 1}, []int{x, y + 1})
					}
				} else {
					if hexMap[x-1][y] == kind && hexMap[x][y+1] == kind {
						matched = append(matched, []int{x, y}, []int{x - 1, y}, []int{x, y + 1})
					}
				}
			}
			// Check 5
			if x < 9 {
				if x%2 == 0 {
					if hexMap[x+1][y] == kind && hexMap[x+1][y+1] == kind {
						matched = append(matched, []int{x, y}, []int{x + 1, y}, []int{x + 1, y + 1})
					}
				} else if y > 0 {
					if hexMap[x+1][y-1] == kind && hexMap[x+1][y] == kind {
						matched = append(matched, []int{x, y}, []int{x + 1, y - 1}, []int{x + 1, y})
					}
				}
			}
			// Check 6
			if y > 0 && x > 0 {
				if x%2 == 0 {
					if hexMap[x-1][y] == kind && hexMap[x][y-1] == kind {
						matched = append(matched, []int{x, y}, []int{x - 1, y}, []int{x, y - 1})
					}
				} else {
					if hexMap[x-1][y-1] == kind && hexMap[x][y-1] == kind {
						matched = append(matched, []int{x, y}, []int{x - 1, y - 1}, []int{x, y - 1})
					}
				}
			}
		}
	}
	var rt [][]int
	for _, v := range matched {
		found := false
		for _, v2 := range rt {
			if v2[0] == v[0] && v2[1] == v[1] {
				found = true
				break
			}
		}
		if !found {
			rt = append(rt, v)
		}
	}
	return rt
}
