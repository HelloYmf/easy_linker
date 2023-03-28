package linker

import "github.com/HelloYmf/elf_linker/pkg/file/elf_file"

type InputElfSymbol struct {
	MparentFile   *InputElfObj     // 所属文件
	MinputSection *InputElfSection // 所属的section
	Mname         string           // 名字
	Mvalue        int
	Msymndx       int32 // SectionHader数组中索引
}

func NewElfInputSymbol(name string) *InputElfSymbol {
	retsym := &InputElfSymbol{
		Mname: name,
	}
	return retsym
}

func GetElfSymbolByName(ctx *LinkContext, name string) *InputElfSymbol {
	if sym, ok := ctx.MsymMap[name]; ok {
		return sym
	}
	ctx.MsymMap[name] = NewElfInputSymbol(name)
	return ctx.MsymMap[name]
}

func (isym *InputElfSymbol) SetInputSection(is *InputElfSection) {
	isym.MinputSection = is
}

func (is *InputElfSymbol) GetSymbolStruct() *elf_file.ElfSymbol {
	if is.Msymndx < int32(len(is.MparentFile.MobjFile.MsectionHdr)) {
		return &is.MparentFile.MobjFile.MsymTable[is.Msymndx]
	}
	return nil
}

func (is *InputElfSymbol) ClearSysmbol() {
	is.MparentFile = nil
	is.MinputSection = nil
	is.Msymndx = -1
}
