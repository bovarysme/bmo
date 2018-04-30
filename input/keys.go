package input

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Event int

const (
	None Event = iota
	Quit
)

type Keys interface {
	Read() Event
}

type SDLKeys struct {
	joypad *Joypad
}

func NewSDLKeys(joypad *Joypad) Keys {
	return &SDLKeys{
		joypad: joypad,
	}
}

func (s *SDLKeys) Read() Event {
	for {
		event := sdl.PollEvent()
		if event == nil {
			break
		}

		switch t := event.(type) {
		case *sdl.QuitEvent:
			return Quit
		case *sdl.KeyDownEvent:
			sym := t.Keysym.Sym
			if sym == sdl.K_q {
				return Quit
			}

			key, ok := s.getKey(sym)
			if ok {
				s.joypad.SetKey(key)
			}
		case *sdl.KeyUpEvent:
			key, ok := s.getKey(t.Keysym.Sym)
			if ok {
				s.joypad.ResetKey(key)
			}
		}
	}

	return None
}

func (s *SDLKeys) getKey(sym sdl.Keycode) (byte, bool) {
	var value byte
	var ok bool = true

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
	default:
		ok = false
	}

	return value, ok
}
