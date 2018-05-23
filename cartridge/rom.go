package cartridge

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
