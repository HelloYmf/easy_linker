package linker

import (
	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
)

type InputElfSection struct {
	MparentFile *elf_file.ElfObjFile // 所属文件
	Mcontents   []byte               // 内部数据
	Mshndx      uint32               // SectionHeader数组内下标
}

func NewElfInputSection(objfil elf_file.ElfObjFile, shndx uint32) *InputElfSection {
	rets := &InputElfSection{
		MparentFile: &objfil,
		Mshndx:      shndx,
	}

	shdr := rets.GetSectionHdr()
	rets.Mcontents = objfil.Mfile.Contents[shdr.Offset : shdr.Offset+shdr.Size]

	return rets
}

func (is *InputElfSection) GetSectionHdr() *elf_file.ElfSectionHdr {
	if is.Mshndx < uint32(len(is.MparentFile.MsectionHdr)) {
		return &is.MparentFile.MsectionHdr[is.Mshndx]
	}
	return nil
}

func (is *InputElfSection) GetParentName() string {
	return is.MparentFile.GetSectionName(is.GetSectionHdr().Name)
}
