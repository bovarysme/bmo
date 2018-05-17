package cartridge

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	headerStart = 0x134
	headerEnd   = 0X150
)

const (
	romBanking byte = iota
	ramBanking
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
	Save() error
}

func NewCartridge(romPath string) (Cartridge, error) {
	rom, err := ioutil.ReadFile(romPath)
	if err != nil {
		return nil, err
	}
	log.Printf("ROM size: %d bytes\n", len(rom))

	if len(rom) < headerEnd {
		return nil, errors.New("Invalid ROM size")
	}

	header, err := NewHeader(rom[headerStart:headerEnd])
	if err != nil {
		return nil, err
	}
	log.Printf("ROM type: %#x\n", header.Type)

	var cartridge Cartridge

	switch header.Type {
	case 0x00:
		cartridge = &ROM{
			rom: rom,
		}
	case 0x01, 0x02, 0x03:
		cartridge = NewMBC1(rom, header.RAMSize)
	case 0x11, 0x12, 0x13:
		cartridge = NewMBC3(rom, header.RAMSize)
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

func (r *ROM) Save() error {
	return nil
}

func getRAMInfo(ramType byte) (int, int) {
	var banks, size int

	switch ramType {
	case 0:
		banks, size = 0, 0
	case 1:
		banks, size = 1, 2048
	case 2:
		banks, size = 1, 8192
	case 3:
		banks, size = 4, 8192
	case 4:
		banks, size = 16, 8192
	case 5:
		banks, size = 8, 8192
	}

	return banks, size
}

func initRAM(ramType byte) [][]byte {
	banks, size := getRAMInfo(ramType)
	if size == 0 {
		return nil
	}

	ram := make([][]byte, banks)
	for i := range ram {
		ram[i] = make([]byte, size)
	}

	return ram
}
