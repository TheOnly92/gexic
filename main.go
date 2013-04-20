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

var hexTex, wallpaper, starTex gl.Texture

var hexMap [10][9]int
var currExX, currExY int
var rotate float32
var timesToRotate int
var currentMatches [][]int
var starMatches [][]int
var starScale float32
var starAlpha float32
var currStarCenter []int
var scale float32
var animateFall []*freefall
var fallticks int
var mouseLock bool

const (
	HEX_WIDTH  int = 44
	HEX_HEIGHT     = 40
)

type freefall struct {
	x, y    int
	targetY int
	accel   float64
}

func genHexMap() {
	hexMap = [10][9]int{
		[9]int{0, 2, 0, 0, 5, 1, 1, 1, -1},
		[9]int{1, 1, 3, 5, 2, 1, 4, 3, 2},
		[9]int{5, 4, 2, 0, 3, 1, 1, 0, -1},
		[9]int{3, 4, 1, 3, 2, 5, 2, 3, 4},
		[9]int{4, 3, 3, 5, 3, 4, 1, 5, -1},
		[9]int{1, 3, 2, 1, 2, 3, 1, 4, 1},
		[9]int{4, 5, 6, 5, 1, 5, 3, 6, -1},
		[9]int{0, 5, 4, 3, 4, 3, 0, 2, 3},
		[9]int{3, 2, 4, 5, 2, 5, 0, 4, -1},
		[9]int{0, 2, 5, 0, 0, 2, 2, 4, 5}}
	return
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
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
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
	// gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.DECAL)
	gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.PushMatrix()
	gl.Translatef(80, 80, 0)
	for x := 0; x < 10; x++ {
		maxy := 8
		topy := 17
		if x%2 == 1 {
			maxy = 9
			topy = 0
		}
		starty := 0
		for y := starty; y < maxy; y++ {
			if y == currExY && x == currExX || currExX%2 == 0 && (x == currExX+1 && y == currExY || x == currExX+1 && y == currExY+1) || currExX%2 == 1 && (x == currExX+1 && y == currExY || x == currExX+1 && y == currExY-1) {
				continue
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
			drawHex(x*30, topy+y*36, hexMap[x][y], 1)
		}
	}
	gl.PopMatrix()
	if len(currentMatches) > 0 || len(starMatches) > 0 {
		mouseLock = true
		if len(currentMatches) > 0 && scale > 0 {
			scale -= 0.1
			for _, v := range currentMatches {
				gl.PushMatrix()
				topy := 17
				if v[0]%2 == 1 {
					topy = 0
				}
				gl.Translatef(float32(v[0]*30+102), float32(v[1]*36+topy+94), 0)
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
			topy := 17
			pm := 0
			if currStarCenter[0]%2 == 1 {
				topy = 0
				pm = -1
			}
			gl.Translatef(float32(currStarCenter[0]*30+102), float32(currStarCenter[1]*36+topy+94), 0)
			drawHex(-22, -14, 6, 1)
			gl.PopMatrix()
			gl.PushMatrix()
			gl.Translatef(float32(currStarCenter[0]*30+102), float32(currStarCenter[1]*36+topy+94), 0)
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
					topy := 17
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
						gl.Translatef(float32(x*30+102), float32(y*36+topy+94), 0)
						gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
						drawHex(-22, -14, hexMap[x][y], 1)
						gl.PopMatrix()
					}
				}
				fallticks++
			} else {
				stillFalling := 0
				for _, v := range animateFall {
					topy := 17
					if v.x%2 == 1 {
						topy = 0
					}
					newy := v.accel * math.Pow(float64(fallticks), 2) / 2
					gl.PushMatrix()
					gl.Translatef(float32(v.x*30+102), float32(math.Min(float64(v.y*36+topy+94)+newy, float64(v.targetY*36+topy+94))), 0)
					gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
					drawHex(-22, -14, hexMap[v.x][v.y], 1)
					gl.PopMatrix()
					if float64(v.y*36+topy+94)+newy < float64(v.targetY*36+topy+94) {
						stillFalling++
					}
				}
				fallticks++
				if stillFalling == 0 {
					starScale = 1
					starAlpha = 0
					removeHexAndGenNew(currentMatches)
					currentMatches = checkHexMap()
					makeStarAndGenNew(starMatches)
					starMatches = [][]int{}
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
		topy := 17
		if currExX%2 == 1 {
			topy = 0
		}
		gl.Translatef(float32(currExX*30+HEX_WIDTH+80), float32(currExY*36+topy+20+80), 0)
		gl.Scalef(1.3, 1.3, 1)
		gl.Rotatef(rotate, 0, 0, 1)
		if currExX%2 == 0 {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			drawHex(-12, -36, hexMap[currExX+1][currExY], 1)
		} else {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			drawHex(-12, -36, hexMap[currExX+1][currExY-1], 1)
		}
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		drawHex(-44, -19, hexMap[currExX][currExY], 1)
		if currExX%2 == 0 {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			drawHex(-12, -2, hexMap[currExX+1][currExY+1], 1)
		} else {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			drawHex(-12, -2, hexMap[currExX+1][currExY], 1)
		}
		gl.PopMatrix()
		if rotate < 120 {
			rotate += 12
		} else {
			if currExX%2 == 0 {
				hexMap[currExX][currExY], hexMap[currExX+1][currExY], hexMap[currExX+1][currExY+1] = hexMap[currExX+1][currExY+1], hexMap[currExX][currExY], hexMap[currExX+1][currExY]
			} else {
				hexMap[currExX][currExY], hexMap[currExX+1][currExY-1], hexMap[currExX+1][currExY] = hexMap[currExX+1][currExY], hexMap[currExX][currExY], hexMap[currExX+1][currExY-1]
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
				} else {
					if timesToRotate == 0 {
						currExX = -1
						currExY = -1
						rotate = 0
						timesToRotate = 0
						mouseLock = false
						fmt.Println("Mouse unlocked 3")
					}
					rotate = 0
					timesToRotate--
				}
			}
		}
	}
	gl.Flush()
	gl.Disable(gl.TEXTURE_2D)
	gl.Disable(gl.BLEND)
}

func main() {
	sys := Make()
	sys.Startup()
	defer sys.Shutdown()
	// InitQueue()

	sys.CreateWindow(1024, 768, "Gexic")
	gl.ClearColor(0., 0.2, 0.4, 0.)
	initGL()

	glfw.SetMouseButtonCallback(mouseButtonCallback)
	glfw.SetCharCallback(charCallback)

	// PurgeQueue()
	genHexMap()
	for matches := checkHexMap(); len(matches) > 0; matches = checkHexMap() {
		removeHexAndGenNew(matches)
	}
	currExX = -1
	currExY = -1

	for sys.CheckExitMainLoop() {
		timer := time.NewTimer(time.Second / 30)
		renderHexMap()
		sys.Refresh()
		<-timer.C
		PurgeQueue()
	}
}

func initGL() {
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, 1024, 768, 0, -1, 1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Enable(gl.DEPTH_TEST)

	hexTex = initTexture("hex3c", HEX_WIDTH, HEX_HEIGHT)
	wallpaper = initTexture("wallpaper-2594238", 1024, 768)
	starTex = initTexture("hexstar", HEX_WIDTH, HEX_HEIGHT)
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
			if (filename == "hex3c" || filename == "hexstar") && r == 0 && g == 0 && b == 0 {
				a = 0
			}
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
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
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
			g = 1
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

func mouseButtonCallback(button, state int) {
	if currExX != -1 || currExY != -1 || mouseLock {
		return
	}
	x, y := glfw.MousePos()

	if state == glfw.KeyPress {
		switch button {
		case glfw.MouseLeft:
			// fmt.Println(x, y)
			currExX = int(math.Floor((float64(x) - 80) / 30))
			currExY = int(math.Floor((float64(y) - 80 - 17) / 36))
			if currExX%2 == 1 {
				currExY = (y - 80) / 36
			}
			if currExX > 9 || currExY > 8 || currExX < 0 || currExY < 0 {
				currExX = -1
				currExY = -1
				return
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
