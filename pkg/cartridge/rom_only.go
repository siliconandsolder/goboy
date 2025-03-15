package cartridge

type RomOnly struct{}

func (m *RomOnly) Read(addr uint16) (uint32, bool) {
	return uint32(addr), true
}

func (m *RomOnly) Write(addr uint16, data byte) (uint32, bool) {
	return uint32(addr), true
}
