package linker

import (
	"bytes"
	"debug/elf"
	"encoding/binary"

	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type ElfOutputEhdr struct {
	ElfChunk
}

func NewElfOutputEhdr() *ElfOutputEhdr {
	return &ElfOutputEhdr{ElfChunk{
		Mhdr: elf_file.ElfSectionHdr{
			Flags:     uint64(elf.SHF_ALLOC),
			Size:      uint64(elf_file.ElfHdrSize),
			AddrAlign: 8,
		},
	},
	}
}

func getFlags(ctx *LinkContext) uint32 {
	if len(ctx.MobjFileList) <= 0 {
		utils.FatalExit("wrong obj files num.")
	}
	flags := ctx.MobjFileList[0].GetEhdr().Flags
	for _, obj := range ctx.MobjFileList[1:] {
		if obj == ctx.MinternalObj {
			continue
		}
		// EF_RISVC_RVC == 1
		if obj.GetEhdr().Flags&1 != 0 {
			flags |= 1
			break
		}
	}
	return flags
}

func (o *ElfOutputEhdr) CopyBuf(ctx *LinkContext) {
	ehdr := elf_file.ElfHdr{}
	copy(ehdr.Ident[:], "\177ELF")
	// 这里写死了64位
	ehdr.Ident[elf.EI_CLASS] = uint8(elf.ELFCLASS64)
	ehdr.Ident[elf.EI_DATA] = uint8(elf.ELFDATA2LSB)
	ehdr.Ident[elf.EI_VERSION] = uint8(elf.EV_CURRENT)
	ehdr.Ident[elf.EI_OSABI] = 0
	ehdr.Ident[elf.EI_ABIVERSION] = 0
	ehdr.Type = uint16(elf.ET_EXEC)
	// 这里写死了RISCV
	ehdr.Machine = uint16(elf.EM_RISCV)
	ehdr.Version = uint32(elf.EV_CURRENT)
	ehdr.Entry = GetEntryAddress(ctx)
	ehdr.PhOff = ctx.MoutPHdr.Mhdr.Offset
	ehdr.ShOff = ctx.MoutSHdr.Mhdr.Offset
	ehdr.Flags = getFlags(ctx)
	ehdr.EhSize = uint16(elf_file.ElfHdrSize)
	ehdr.PhEntSize = uint16(elf_file.ElfProgramHdrSize)
	ehdr.PhNum = uint16(ctx.MoutPHdr.Mhdr.Size) / uint16(elf_file.ElfProgramHdrSize)
	ehdr.ShEntSize = uint16(elf_file.ElfSectionHdrSize)
	ehdr.ShNum = uint16(ctx.MoutSHdr.Mhdr.Size) / uint16(elf_file.ElfSectionHdrSize)

	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, ehdr)
	utils.MustNoErr(err)
	copy(ctx.Mbuf, buf.Bytes())
}

func GetEntryAddress(ctx *LinkContext) uint64 {
	for _, osec := range ctx.MoutputSections {
		if osec.Mname == ".text" {
			return osec.Mhdr.Addr
		}
	}
	return 0
}
