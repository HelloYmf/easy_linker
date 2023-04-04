package linker

import (
	"bytes"
	"debug/elf"
	"fmt"
	"math"

	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type InputElfObj struct {
	MobjFile           elf_file.ElfObjFile    // 基类obj文件--提供符号结构数组
	MinputSections     []*InputElfSection     // InputSection数组，对基础SectionHdr的封装
	Msymtabndxdata     []uint32               // 符号表索引表数据
	MisUsed            bool                   // 表示当前obj文件是否被实际使用了
	MallSymbols        []*InputElfSymbol      // 当前文件中所有用到的符号
	MlocalSymbols      []InputElfSymbol       // 当前文件中所有局部符号
	MmergeableSections []*ElfMergeableSection // 当前文件中所有可合并的节
}

func NewElfInputObj(ctx *LinkContext, f *elf_file.ElfObjFile) *InputElfObj {

	reto := &InputElfObj{MobjFile: *f}
	reto.MisUsed = false

	// 解析符号表--初始化Symbol结构数组
	reto.MobjFile.PraseSymbolTable()
	// 初始化sections--继续封装一层section header名字为InputSection，
	// 每个InputSection可以绑定所属的obj文件，
	// 以及根据name、type、flag等信息生成一个OutputSection
	// 一个obj文件对应多个InputSection，多个InputSection可能对应一个OutputSection，并且OutputSection存在于全局context结构中，并且如果name、type、flag一致就认为是同一个
	reto.InitialzeSections(ctx)
	// 初始化符号--初始化局部符号和全局符号
	// InputObjFile中一个所有符号的指针表，以及只属于当前obj文件的局部符号表，全局符号数组在context结构中使用map唯一保存，因为符号不能重定义
	// 对于局部符号只需要将符号的所属obj文件设置为自己，并插入数组中即可
	// 对于全局符号，需要检查map文件中是否存在，存在就直接返回指针。不存在则创建并更新map，后续解析其他文件时再补充parent信息
	reto.InitialzeSymbols(ctx)
	// 初始化可合并的节--弃用InputSection，在context中建立合并section map
	// 根据section hdr->flag判断可合并的section，如果发现了，初始化一个MergeableSection
	// 此时这个section不再受InputSection，而是MergeableSection,InputSection对应的IsUsed置false
	// 每个MergeableSection归MergedSection管理，多对一的关系
	// 并且这个MergedSection也跟全局符号一样由context结构维护
	// 每个MergeableSection的内部数据由多个大小为EntSize的fragment块组成
	reto.InitialzeMergeableSections(ctx)
	// 跳过异常节，将异常InputSection的IsUsed置false
	reto.SkipEhFrameSections()

	return reto
}

func (io *InputElfObj) InitialzeSections(ctx *LinkContext) {
	secnum := len(io.MobjFile.MsectionHdr)
	io.MinputSections = make([]*InputElfSection, secnum)
	for i := 0; i < secnum; i++ {
		shdr := &io.MobjFile.MsectionHdr[i]
		switch elf.SectionType(shdr.Type) {
		case elf.SHT_GROUP, elf.SHT_SYMTAB, elf.SHT_STRTAB, elf.SHT_REL, elf.SHT_RELA, elf.SHT_NULL:
		case elf.SHT_SYMTAB_SHNDX:
			// 填充索引数据节
			io.GetSymtabShndxSecdata(shdr)
		default: // 需要填充进可执行文件的sections
			name := io.MobjFile.GetSectionName(shdr.Name)
			io.MinputSections[i] = NewElfInputSection(ctx, name, io, uint32(i))
		}
	}

	for i := 0; i < len(io.MinputSections); i++ {
		shdr := &io.MobjFile.MsectionHdr[i]
		if shdr.Type != uint32(elf.SHT_RELA) {
			continue
		}
		if shdr.Info >= uint32(len(io.MinputSections)) {
			utils.FatalExit("Wrong Relocation Info.")
		}

		if target := io.MinputSections[shdr.Info]; target != nil {
			if target.MrelSecIdx != math.MaxUint32 {
				utils.FatalExit("Wrong MrelSecIdx.")
			}
			target.MrelSecIdx = uint32(i)
		}
	}
}

func (io *InputElfObj) InitialzeSymbols(ctx *LinkContext) {
	if io.MobjFile.MsymTable == nil {
		// 这里不能直接退出，因为有的库obj文件没有符号
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
		// 局部符号的所属obj文件都是自己
		local_sym.MparentFile = io
		local_sym.Mvalue = int(symhdr.Val)
		local_sym.Msymndx = int32(i)

		if !symhdr.IsAbs() {
			index := io.GetSymtabShndx(symhdr, uint64(i))
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
		symhdr := io.MobjFile.MsymTable[i]
		name := io.MobjFile.GetSymbolName(symhdr.Name)
		// 检查全局符号map，如果已经存在，直接返回
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

func (io *InputElfObj) GetSymtabShndx(sym elf_file.ElfSymbol, ndx uint64) int64 {
	if ndx > uint64(len(io.MobjFile.MsymTable)) {
		utils.FatalExit("range of MsymTable.")
	}
	if sym.Shndx == uint16(elf.SHN_XINDEX) {
		return int64(io.Msymtabndxdata[ndx])
	}
	return int64(sym.Shndx)
}

// 收集全部的已经定义的全局符号
func (io *InputElfObj) DealGlobalSymbols() {
	for i := io.MobjFile.MglobalSymndx; i < uint32(len(io.MobjFile.MsymTable)); i++ {
		sym := io.MallSymbols[i]
		symhdr := &io.MobjFile.MsymTable[i]

		// 跳过未定义的全局符号
		if symhdr.IsUndef() {
			continue
		}

		var isec *InputElfSection = nil
		if !symhdr.IsAbs() {
			isec = io.GetSection(*symhdr, uint64(i))
			if isec == nil {
				continue
			}
		}

		// 如果此时parent obj文件为空，表示这是一个全局符号，对于当前文件来说不是一个未定义符号，此时当前文件就是所属文件
		if sym.MparentFile == nil {
			sym.MparentFile = io
			sym.SetInputSection(isec)
			sym.Mvalue = int(symhdr.Val)
			sym.Msymndx = int32(i)
		}

	}
}

func (io *InputElfObj) GetSection(sym elf_file.ElfSymbol, ndx uint64) *InputElfSection {
	idx := io.GetSymtabShndx(sym, ndx)
	if idx == -1 {
		return nil
	}
	return io.MinputSections[idx]
}

func (io *InputElfObj) MarkLiveObjs(feeder func(*InputElfObj)) {
	if !io.MisUsed {
		return
	}
	for i := io.MobjFile.MglobalSymndx; i < uint32(len(io.MobjFile.MsymTable)); i++ {
		sym := io.MallSymbols[i]
		symhdr := &io.MobjFile.MsymTable[i]
		if sym.MparentFile == nil {
			continue
		}

		// 经过上面的所有obj文件解析，已经在全局context列表中进行了更新，每个全局符号都有了所属的parent
		if symhdr.IsUndef() && !sym.MparentFile.MisUsed {
			// 将使用到了但是还没激活的parent obj file激活
			sym.MparentFile.MisUsed = true
			// 递归处理全部符号之间的关系
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

func (io *InputElfObj) InitialzeMergeableSections(ctx *LinkContext) {
	io.MmergeableSections = make([]*ElfMergeableSection, len(io.MinputSections))
	for i := 0; i < len(io.MinputSections); i++ {
		sec := io.MinputSections[i]
		if sec != nil && sec.MisUserd && uint32(sec.MparentFile.MinputSections[i].GetSectionHdr().Flags)&uint32(elf.SHF_MERGE) != 0 {
			// 根据InputSection获得一个可合并的section，并建立全局map
			io.MmergeableSections[i] = NewMergeableSections(ctx, sec)
			// 将处理过的InputSection设置为未使用，后面只需要操作可合并section即可
			sec.MisUserd = false
		}
	}
}

func findStringNil(data []byte, endSize int) int {
	if endSize == 1 {
		return bytes.Index(data, []byte{0})
	}
	for i := 0; i <= len(data)-endSize; i += endSize {
		bs := data[i : i+endSize]
		if utils.AllZeros(bs) {
			return i
		}
	}
	return -1
}

func NewMergeableSections(ctx *LinkContext, isec *InputElfSection) *ElfMergeableSection {
	ms := &ElfMergeableSection{}
	hdr := isec.GetSectionHdr()

	// 根据section名字以及类型去全局合并后map中获取，如果获取不到就新增一个
	ms.Mparent = GetMergedSectionInstance(ctx, isec.GetSectionName(), uint64(hdr.Type), hdr.Flags)
	ms.Mp2Align = isec.Mp2Align

	data := isec.Mcontents
	offset := 0

	if hdr.Flags&uint64(elf.SHF_STRINGS) != 0 {
		for len(data) > 0 {
			end := findStringNil(data, int(hdr.EntSize))
			if end == -1 {
				utils.FatalExit("string not end.")
			}
			subdata := data[:end+int(hdr.EntSize)]
			data = data[end+int(hdr.EntSize):]
			ms.Moridata = append(ms.Moridata, string(subdata))
			ms.MblockOffset = append(ms.MblockOffset, uint32(offset))
			offset += int(hdr.EntSize)
		}
	} else {
		if len(data)%int(hdr.EntSize) != 0 {
			utils.FatalExit("section data size wrong.")
		}
		for len(data) > 0 {
			subdata := data[:hdr.EntSize]
			data = data[hdr.EntSize:]
			ms.Moridata = append(ms.Moridata, string(subdata))
			ms.MblockOffset = append(ms.MblockOffset, uint32(offset))
			offset += int(hdr.EntSize)
		}
	}
	return ms
}

func (io *InputElfObj) RegisterSectionPieces() {
	for _, ms := range io.MmergeableSections {
		if ms == nil {
			continue
		}
		ms.Mblock = make([]*ElfSectionBlock, 0, len(ms.Moridata))
		for i := 0; i < len(ms.Moridata); i++ {
			ms.Mblock = append(ms.Mblock, ms.Mparent.Insert(ms.Moridata[i], ms.Mp2Align))
		}
	}

	for i := 1; i < len(io.MobjFile.MsymTable); i++ {

		sym := io.MallSymbols[i]
		symhdr := &io.MobjFile.MsymTable[i]

		if symhdr.IsAbs() || symhdr.IsUndef() || symhdr.IsCommon() {
			continue
		}

		idx := io.GetSymtabShndx(*symhdr, uint64(i))
		m := io.MmergeableSections[idx]
		if m == nil {
			continue
		}

		block, off := m.GetBlock(uint32(symhdr.Val))
		if block == nil {
			utils.FatalExit("wrong block.")
		}
		sym.SetSectionBlock(block)
		sym.Mvalue = int(off)
	}
}

func (f *InputElfObj) GetEhdr() elf_file.ElfHdr {
	return utils.BinRead[elf_file.ElfHdr](f.MobjFile.Mfile.Contents)
}

// 过滤.eh_frame，这个是处理异常的
func (f *InputElfObj) SkipEhFrameSections() {
	for _, isec := range f.MinputSections {
		if isec != nil && isec.MisUserd && isec.GetSectionName() == ".eh_frame" {
			isec.MisUserd = false
		}
	}
}

func (f *InputElfObj) ScanRelocations() {
	for _, isec := range f.MinputSections {
		if isec != nil && isec.MisUserd &&
			isec.GetSectionHdr().Flags&uint64(elf.SHF_ALLOC) != 0 {
			isec.ScanAllRelocations()
		}
	}
}

func (f *InputElfObj) GetBytesFromShdr(s *elf_file.ElfSectionHdr) []byte {
	end := s.Offset + s.Size
	if uint64(len(f.MobjFile.Mfile.Contents)) < end {
		utils.FatalExit(
			fmt.Sprintf("section header is out of range: %d", s.Offset))
	}
	return f.MobjFile.Mfile.Contents[s.Offset:end]
}
