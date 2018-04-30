package screen

import (
	"github.com/bovarysme/bmo/ppu"

	"github.com/veandco/go-sdl2/sdl"
)

type Screen interface {
	Render(pixels []byte) error
	Shutdown()
}

type SDLScreen struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
}

func NewSDLScreen(screenScale int) (*SDLScreen, error) {
	err := sdl.Init(sdl.INIT_VIDEO)
	if err != nil {
		return nil, err
	}

	screen := &SDLScreen{}

	window, err := sdl.CreateWindow("BMO", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		ppu.ScreenWidth*screenScale, ppu.ScreenHeight*screenScale, sdl.WINDOW_SHOWN)
	if err != nil {
		screen.Shutdown()
		return nil, err
	}
	screen.window = window

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		screen.Shutdown()
		return nil, err
	}
	screen.renderer = renderer

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGB888, sdl.TEXTUREACCESS_STREAMING,
		ppu.ScreenWidth, ppu.ScreenHeight)
	if err != nil {
		screen.Shutdown()
		return nil, err
	}
	screen.texture = texture

	err = renderer.Clear()
	if err != nil {
		screen.Shutdown()
		return nil, err
	}

	return screen, nil
}

func (s *SDLScreen) Render(pixels []byte) error {
	err := s.texture.Update(nil, pixels, ppu.Pitch)
	if err != nil {
		s.Shutdown()
		return err
	}

	err = s.renderer.Copy(s.texture, nil, nil)
	if err != nil {
		s.Shutdown()
		return err
	}

	s.renderer.Present()

	return nil
}

func (s *SDLScreen) Shutdown() {
	if s.texture != nil {
		s.texture.Destroy()
	}

	if s.renderer != nil {
		s.renderer.Destroy()
	}

	if s.window != nil {
		s.window.Destroy()
	}

	sdl.Quit()
}
