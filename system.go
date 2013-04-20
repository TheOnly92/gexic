package main

import (
	"fmt"
	"github.com/go-gl/gl"
	"github.com/go-gl/glfw"
	"os"
)

type System struct {
}

func Make() *System {
	return &System{}
}

func (s *System) Startup() {
	if err := glfw.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}

	// glfw.OpenWindowHint(glfw.FsaaSamples, 4)
	// glfw.OpenWindowHint(glfw.OpenGLVersionMajor, 3)
	// glfw.OpenWindowHint(glfw.OpenGLVersionMinor, 2)
	// glfw.OpenWindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	InitQueue()
}

func (s *System) CreateWindow(width, height int, title string) {
	if err := glfw.OpenWindow(width, height, 0, 0, 0, 8, 32, 0, glfw.Windowed); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}

	gl.Init()       // Can't find gl.GLEW_OK or any variation, not sure how to check if this worked
	CheckGLErrors() // Ignore error

	glfw.SetWindowTitle(title)

	glfw.Enable(glfw.StickyKeys)

	gl.ClearColor(0., 0., 0., 0.)

	gl.Clear(gl.COLOR_BUFFER_BIT)
}

func (s *System) Shutdown() {
	glfw.Terminate()
}

func (s *System) CheckExitMainLoop() bool {
	return (glfw.Key(glfw.KeyEsc) != glfw.KeyPress && glfw.WindowParam(glfw.Opened) == gl.TRUE)
}

func (s *System) Refresh() {
	// gl.Clear(gl.COLOR_BUFFER_BIT)
	glfw.SwapBuffers()
}

func CheckGLErrors() {
	glerror := gl.GetError()
	if glerror == gl.NO_ERROR {
		return
	}

	fmt.Printf("gl.GetError() reports")
	for glerror != gl.NO_ERROR {
		fmt.Printf(" ")
		switch glerror {
		case gl.INVALID_ENUM:
			fmt.Printf("GL_INVALID_ENUM")
		case gl.INVALID_VALUE:
			fmt.Printf("GL_INVALID_VALUE")
		case gl.INVALID_OPERATION:
			fmt.Printf("GL_INVALID_OPERATION")
		case gl.STACK_OVERFLOW:
			fmt.Printf("GL_STACK_OVERFLOW")
		case gl.STACK_UNDERFLOW:
			fmt.Printf("GL_STACK_UNDERFLOW")
		case gl.TABLE_TOO_LARGE:
			fmt.Printf("GL_TABLE_TOO_LARGE")
		case gl.OUT_OF_MEMORY:
			fmt.Printf("GL_OUT_OF_MEMORY")
		default:
			fmt.Printf("%d", glerror)
		}
		glerror = gl.GetError()
	}
	fmt.Printf("\n")
}
