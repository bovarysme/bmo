# bmo

A Game Boy emulator written in Go.

## Progress

- Cartridges (ROM only, MBC1)
- CPU (passes Blargg's CPU instruction test)

![Blargg's CPU instruction test passed](docs/cpu_instr.png)

- Interrupt Controller
- PPU (background and window rendering, scrolling)

## Requirements

- SDL2 (`libsdl2-dev` on Debian-based distributions)

## Usage

- Copy the DMG bootrom to `roms/bootrom.gb` (or specify its path with the
  `-bootrom` flag)
- Compile and run the emulator:

```
$ dep ensure
$ go build
$ ./bmo -rom <path to the ROM file>
```

## References

- [Gameboy CPU (LR35902) instruction set](http://www.pastraiser.com/cpu/gameboy/gameboy_opcodes.html)
- Game Boy Programming Manual
- [Pan Docs](http://gbdev.gg8.se/wiki/articles/Pan_Docs)
- [The Ultimate Game Boy Talk](https://www.youtube.com/watch?v=HyzD8pNlpwI)

## License

This project is licensed under the terms of the MIT license.
