package elf_file

import "unsafe"

const ElfHdrSize = unsafe.Sizeof(ElfHdr{})
const SectionHdrSize = unsafe.Sizeof(SectionHdr{})

type ElfHdr struct {
	Ident     [16]uint8
	Type      uint16
	Machine   uint16
	Version   uint32
	Entry     uint64
	PhOff     uint64
	ShOff     uint64
	Flags     uint32
	EhSize    uint16
	PhEntSize uint16
	PhNum     uint16
	ShEntSize uint16
	ShNum     uint16
	ShStrndx  uint16
}

type SectionHdr struct {
	Name      uint32
	Type      uint32
	Flags     uint64
	Addr      uint64
	Offset    uint64
	Size      uint64
	Link      uint32
	Info      uint32
	AddrAlign uint64
	EntSize   uint64
}
