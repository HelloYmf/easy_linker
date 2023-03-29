package linker

import "github.com/HelloYmf/elf_linker/pkg/file/elf_file"

// 输出section的基类
type ELfThunk struct {
	Mname string
	Mhdr  elf_file.ElfSectionHdr
}

func NewThunk() ELfThunk {
	return ELfThunk{Mhdr: elf_file.ElfSectionHdr{AddrAlign: 1}}
}
