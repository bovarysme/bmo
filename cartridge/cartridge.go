package cartridge

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	headerStart = 0x134
	headerEnd   = 0X150
)

type UnknownCartridgeTypeError struct {
	cartridgeType byte
}

func (e UnknownCartridgeTypeError) Error() string {
	return fmt.Sprintf("Unknown cartridge type: %#x", e.cartridgeType)
}

type Header struct {
	Title            [16]byte
	NewLicenseeCode  [2]byte
	SGBFlag          byte
	Type             byte
	ROMSize          byte
	RAMSize          byte
	DestinationCode  byte
	OldLicenseeCode  byte
	ROMVersionNumber byte
	HeaderChecksum   byte
	GlobalChecksum   [2]byte
}

func NewHeader(data []byte) (*Header, error) {
	header := &Header{}

	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, header)
	if err != nil {
		return nil, err
	}

	return header, nil
}

type Cartridge interface {
	ReadByte(address uint16) byte
	WriteByte(address uint16, value byte)
}

func NewCartridge(rom []byte) (Cartridge, error) {
	if len(rom) < headerEnd {
		return nil, errors.New("Invalid ROM size")
	}

	header, err := NewHeader(rom[headerStart:headerEnd])
	if err != nil {
		return nil, err
	}

	var cartridge Cartridge

	switch header.Type {
	case 0x00:
		cartridge = &ROM{
			rom: rom,
		}
	default:
		return nil, &UnknownCartridgeTypeError{cartridgeType: header.Type}
	}

	return cartridge, nil
}

type ROM struct {
	rom []byte
}

func (r *ROM) ReadByte(address uint16) byte {
	return r.rom[address]
}

func (r *ROM) WriteByte(address uint16, value byte) {

}
