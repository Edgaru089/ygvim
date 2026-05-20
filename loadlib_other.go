//go:build !windows

package main

import (
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

// loads SDL libraries.
func loadlibrary() {
	sdl.LoadLibrary(sdl.Path())
	ttf.LoadLibrary(ttf.Path())
}

// unloads SDL libraries.
func unloadlibrary() {
	ttf.Quit()
	sdl.Quit()
}
