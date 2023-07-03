package cartridge

// MBC returns real address for ROM or RAM
type MBC interface {
	Read(addr uint16) uint32
	Write(addr uint16, data byte) uint32
}
