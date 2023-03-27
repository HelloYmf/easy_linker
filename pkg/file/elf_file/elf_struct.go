package elf_file

import (
	"bytes"
	"strconv"
	"strings"
	"unsafe"

	"github.com/HelloYmf/elf_linker/pkg/utils"
)

const ElfHdrSize = unsafe.Sizeof(ElfHdr{})
const ElfSectionHdrSize = unsafe.Sizeof(ElfSectionHdr{})
const ElfSymbolSize = unsafe.Sizeof(ElfSymbol{})
const ElfArHdrSize = unsafe.Sizeof(ElfArHeader{})

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

type ElfSectionHdr struct {
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

type ElfSymbol struct {
	Name  uint32
	Info  uint8
	Other uint8
	Shndx uint16
	Val   uint64
	Size  uint64
}

type ElfArHeader struct {
	Name [16]byte
	Date [12]byte
	Uid  [6]byte
	Gid  [6]byte
	Mode [8]byte
	Size [10]byte
	Fmag [2]byte
}

func (ar *ElfArHeader) hdrReadDataSize() int {
	size, err := strconv.Atoi(strings.TrimSpace(string((*ar).Size[:])))
	utils.MustNoErr(err)
	return size
}

func (ar *ElfArHeader) hdrIsStrTab() bool {
	return strings.HasPrefix(string((*ar).Name[:]), "// ")
}

func (ar *ElfArHeader) hdrIsSymTab() bool {
	return strings.HasPrefix(string((*ar).Name[:]), "/ ") || strings.HasPrefix(string((*ar).Name[:]), "/SYM64/ ")
}

func (ar *ElfArHeader) hdrReadName(strtab []byte) string {
	// 名字存在strtab中
	if strings.HasPrefix(string((*ar).Name[:]), "/") {
		// TODO
		start, err := strconv.Atoi(strings.TrimSpace(string((*ar).Name[1:])))
		utils.MustNoErr(err)
		end := start + bytes.Index(strtab[start:], []byte("/\n"))
		return string(strtab[start:end])

	} else {
		end := bytes.Index((*ar).Name[:], []byte("/"))
		if end != -1 {
			return string((*ar).Name[:end])
		}
	}
	return ""
}
