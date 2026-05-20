//go:build windows

package main

import (
	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/bin/binttf"
)

var bin_sdl, bin_ttf interface {
	Unload()
}

// loads SDL libraries.
func loadlibrary() {
	bin_sdl = binsdl.Load()
	bin_ttf = binttf.Load()
}

// unloads SDL libraries.
func unloadlibrary() {
	bin_ttf.Unload()
	bin_sdl.Unload()
}
