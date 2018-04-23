package joypad

import (
	"github.com/veandco/go-sdl2/sdl"
)

func Step(joypad *Joypad) {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.KeyDownEvent:
			key := getKey(t.Keysym.Sym)
			joypad.SetKey(key)
		case *sdl.KeyUpEvent:
			key := getKey(t.Keysym.Sym)
			joypad.ResetKey(key)
		}
	}
}

func getKey(sym sdl.Keycode) byte {
	var value byte

	switch sym {
	case sdl.K_RIGHT:
		value = Right
	case sdl.K_LEFT:
		value = Left
	case sdl.K_UP:
		value = Up
	case sdl.K_DOWN:
		value = Down
	case sdl.K_a:
		value = A
	case sdl.K_s:
		value = B
	case sdl.K_c:
		value = Select
	case sdl.K_v:
		value = Start
	}

	return value
}
