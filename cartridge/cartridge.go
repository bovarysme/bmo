package cartridge

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
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

	ram := initRAM(header.RAMSize)
	if hasBattery(header.Type) {
		err = loadRAM(romPath, ram)
		if err != nil {
			return nil, err
		}
	}

	var cartridge Cartridge

	switch header.Type {
	case 0x00:
		cartridge = &ROM{
			rom: rom,
		}
	case 0x01, 0x02, 0x03:
		cartridge = NewMBC1(header.Type, romPath, rom, ram)
	case 0x11, 0x12, 0x13:
		cartridge = NewMBC3(header.Type, romPath, rom, ram)
	default:
		return nil, &UnknownCartridgeTypeError{cartridgeType: header.Type}
	}

	return cartridge, nil
}

func hasBattery(cartType byte) bool {
	switch cartType {
	case 0x03, 0x13:
		return true
	}

	return false
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

func getRAMPath(path string) string {
	ext := filepath.Ext(path)
	ramPath := strings.TrimSuffix(path, ext) + ".bmo"

	return ramPath
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

func loadRAM(path string, ram [][]byte) error {
	ramPath := getRAMPath(path)

	_, err := os.Stat(ramPath)
	if os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(ramPath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, bank := range ram {
		_, err = file.Read(bank)
		if err != nil {
			return err
		}
	}

	log.Printf("Loaded external RAM from '%s'\n", ramPath)

	return nil
}

func saveRAM(path string, ram [][]byte) error {
	ramPath := getRAMPath(path)

	file, err := os.Create(ramPath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, bank := range ram {
		_, err = file.Write(bank)
		if err != nil {
			return err
		}
	}

	log.Printf("Saved external RAM to '%s'\n", ramPath)

	return nil
}
