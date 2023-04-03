package linker

import "github.com/HelloYmf/elf_linker/pkg/file/elf_file"

const (
	NeedsGotTp uint32 = 1 << 0
)

type InputElfSymbol struct {
	MparentFile *InputElfObj // 所属文件
	Mname       string       // 名字
	Mvalue      int
	Msymndx     int32 // SectionHeader数组中索引
	MgotTpIndx  int32

	MinputSection *InputElfSection // 所属的section
	MsectionBlock *ElfSectionBlock // 所属的block，与Inputsection互斥

	Mflags uint32
}

func NewElfInputSymbol(name string) *InputElfSymbol {
	retsym := &InputElfSymbol{
		Mname:   name,
		Msymndx: -1,
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
	isym.MsectionBlock = nil
}

func (isym *InputElfSymbol) SetSectionBlock(sb *ElfSectionBlock) {
	isym.MsectionBlock = sb
	isym.MinputSection = nil
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

func (is *InputElfSymbol) GetAddr() uint64 {
	if is.MsectionBlock != nil {
		return is.MsectionBlock.GetAddr() + uint64(is.Mvalue)
	}

	if is.MinputSection != nil {
		return is.MinputSection.GetAddr() + uint64(is.Mvalue)
	}

	return uint64(is.Mvalue)
}

func (is *InputElfSymbol) GetGotTpAddr(ctx *LinkContext) uint64 {
	return ctx.MgotSection.Mhdr.Addr + uint64(is.MgotTpIndx)*8
}
