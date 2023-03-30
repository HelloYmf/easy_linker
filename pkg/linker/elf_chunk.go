package linker

import "github.com/HelloYmf/elf_linker/pkg/file/elf_file"

type ElfChunker interface {
	GetName() string
	UpdateSHdr(ctx *LinkContext)
	GetShndx() int64
	GetSHdr() *elf_file.ElfSectionHdr
	CopyBuf(ctx *LinkContext)
}

// 输出块的基类
type ElfChunk struct {
	Mname string
	Mhdr  elf_file.ElfSectionHdr
}

func NewThunk() ElfChunk {
	return ElfChunk{Mhdr: elf_file.ElfSectionHdr{AddrAlign: 1}}
}

func (ec *ElfChunk) GetName() string {
	return ""
}

func (ec *ElfChunk) UpdateSHdr(ctx *LinkContext) {
}

func (ec *ElfChunk) GetShndx() int64 {
	return 0
}

func (ec *ElfChunk) GetSHdr() *elf_file.ElfSectionHdr {
	return &ec.Mhdr
}

func (ec *ElfChunk) CopyBuf(ctx *LinkContext) {
}
