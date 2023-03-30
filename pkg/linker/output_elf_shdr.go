package linker

import (
	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type ElfOutputShdr struct {
	ElfChunk
}

func NewElfOutputShdr() *ElfOutputShdr {
	o := &ElfOutputShdr{ElfChunk: NewThunk()}
	o.Mhdr.AddrAlign = 8
	return o
}

func (s *ElfOutputShdr) UpdateSHdr(ctx *LinkContext) {
	n := uint64(0)
	for _, chunk := range ctx.Mchunks {
		if chunk.GetShndx() > 0 {
			n = uint64(chunk.GetShndx())
		}
	}
	s.Mhdr.Size = (n + 1) * uint64(elf_file.ElfSectionHdrSize)
}

func (s *ElfOutputShdr) CoptBuf(ctx *LinkContext) {
	base := ctx.Mbuf[s.Mhdr.Offset:]
	utils.Write(base, elf_file.ElfSectionHdr{})

	for _, chunk := range ctx.Mchunks {
		if chunk.GetShndx() > 0 {
			utils.Write(base[chunk.GetShndx()*int64(elf_file.ElfSectionHdrSize):],
				*chunk.GetSHdr())
		}
	}
}
