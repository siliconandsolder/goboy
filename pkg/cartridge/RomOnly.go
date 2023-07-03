package cartridge

type RomOnly struct{}

func (m *RomOnly) Read(addr uint16) uint32 {
	return uint32(addr)
}

func (m *RomOnly) Write(addr uint16, data byte) uint32 {
	return uint32(addr)
}
