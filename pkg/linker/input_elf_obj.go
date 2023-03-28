package linker

import (
	"debug/elf"
	"fmt"

	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type InputElfObj struct {
	MobjFile       elf_file.ElfObjFile
	MinputSections []*InputElfSection
	Msymtabndxdata []uint32
	MisUsed        bool

	MallSymbols   []*InputElfSymbol // 当前文件中所有用到的符号
	MlocalSymbols []InputElfSymbol  // 当前文件中所有局部符号
}

func NewElfInputObj(ctx *LinkContext, f *elf_file.ElfObjFile) *InputElfObj {
	reto := &InputElfObj{MobjFile: *f}
	reto.MisUsed = false

	reto.InitialzeSections()
	// 解析符号表
	reto.MobjFile.PraseSymbolTable(true)
	reto.InitialzeSymbols(ctx)

	return reto
}

func (io *InputElfObj) InitialzeSections() {
	secnum := len(io.MobjFile.MsectionHdr)
	io.MinputSections = make([]*InputElfSection, secnum)
	for i := 0; i < secnum; i++ {
		shdr := &io.MobjFile.MsectionHdr[i]
		switch elf.SectionType(shdr.Type) {
		case elf.SHT_GROUP, elf.SHT_SYMTAB, elf.SHT_STRTAB, elf.SHT_REL, elf.SHT_RELA, elf.SHT_NULL:
		case elf.SHT_SYMTAB_SHNDX:
			fmt.Println("SHT_SYMTAB_SHNDX found.")
			// 填充索引数据节
			io.GetSymtabShndxSecdata(shdr)
		default: // 需要填充进可执行文件的sections
			io.MinputSections[i] = NewElfInputSection(io.MobjFile, uint32(i))
		}
	}
}

func (io *InputElfObj) InitialzeSymbols(ctx *LinkContext) {
	if io.MobjFile.MsymTable == nil {
		return
	}

	// 处理全部的局部符号
	io.MlocalSymbols = make([]InputElfSymbol, io.MobjFile.MglobalSymndx)
	for i := 0; i < len(io.MlocalSymbols); i++ {
		io.MlocalSymbols[i] = *NewElfInputSymbol("")
	}
	// 第一个局部符号有特殊意义，暂时不处理，同时将类型设置为Undef，因为undef枚举就是0
	io.MlocalSymbols[0].MparentFile = io
	for i := 1; i < len(io.MlocalSymbols); i++ {
		symhdr := io.MobjFile.MsymTable[i]
		local_sym := &io.MlocalSymbols[i]
		local_sym.Mname = io.MobjFile.GetSymbolName(symhdr.Name)
		local_sym.MparentFile = io
		local_sym.Mvalue = int(symhdr.Val)
		local_sym.Msymndx = int32(i)

		if !symhdr.IsAbs() {
			index := io.GetSymtabShndx(symhdr, i)
			if index == -1 {
				continue
			}
			local_sym.SetInputSection(io.MinputSections[index])
		}
	}
	io.MallSymbols = make([]*InputElfSymbol, len(io.MobjFile.MsymTable))
	for i := 0; i < len(io.MlocalSymbols); i++ {
		io.MallSymbols[i] = &io.MlocalSymbols[i]
	}
	// 处理全部的全局符号
	for i := len(io.MlocalSymbols); i < len(io.MobjFile.MsymTable); i++ {
		symhdr := &io.MobjFile.MsymTable[i]
		name := io.MobjFile.GetSymbolName(symhdr.Name)
		io.MallSymbols[i] = GetElfSymbolByName(ctx, name)
	}
}

func (io *InputElfObj) GetSymtabShndxSecdata(shdr *elf_file.ElfSectionHdr) {
	bs := io.MobjFile.GetSectionData(int64(shdr.Offset))
	num := len(bs) / 4
	for num > 0 {
		io.Msymtabndxdata = append(io.Msymtabndxdata, utils.BinRead[uint32](bs))
		bs = bs[4:]
		num--
	}
}

func (io *InputElfObj) GetSymtabShndx(sym elf_file.ElfSymbol, ndx int) int64 {
	if ndx > len(io.Msymtabndxdata) {
		return -1
	}
	if sym.Shndx == uint16(elf.SHN_XINDEX) {
		return int64(io.Msymtabndxdata[ndx])
	}
	return int64(sym.Shndx)
}

// 收集全部的已经定义的全局符号
func (io *InputElfObj) ResolveSymbols() {
	for i := io.MobjFile.MglobalSymndx; i < uint32(len(io.MobjFile.MsymTable)); i++ {
		sym := io.MallSymbols[i]
		symhdr := &io.MobjFile.MsymTable[i]

		// 跳过未定义的全局符号
		if symhdr.IsUndef() {
			continue
		}

		var isec *InputElfSection
		if symhdr.IsAbs() {
			isec = io.GetSection(*symhdr, int(i))
			if isec == nil {
				continue
			}
		}

		if sym.MparentFile == nil {
			sym.MparentFile = io
			sym.SetInputSection(isec)
			sym.Mvalue = int(symhdr.Val)
			sym.Msymndx = int32(i)
		}

	}
}

func (io *InputElfObj) GetSection(sym elf_file.ElfSymbol, ndx int) *InputElfSection {
	idx := io.GetSymtabShndx(sym, ndx)
	if idx == -1 {
		return nil
	}
	return io.MinputSections[idx]
}

func (io *InputElfObj) MarkLiveObjs(ctx *LinkContext, feeder func(*InputElfObj)) {
	if !io.MisUsed {
		return
	}
	for i := io.MobjFile.MglobalSymndx; i < uint32(len(io.MobjFile.MsymTable)); i++ {
		sym := io.MallSymbols[i]
		symhdr := &io.MobjFile.MsymTable[i]
		if sym.MparentFile == nil {
			continue
		}

		if symhdr.IsUndef() && !sym.MparentFile.MisUsed {
			sym.MparentFile.MisUsed = true
			feeder(sym.MparentFile)
		}

	}
}

func (io *InputElfObj) ClearSymbols() {
	for _, sym := range io.MallSymbols[io.MobjFile.MglobalSymndx:] {
		if sym.MparentFile == io {
			sym.ClearSysmbol()
		}
	}
}
