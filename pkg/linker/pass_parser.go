package linker

import (
	"debug/elf"
	"math"
	"sort"

	"github.com/HelloYmf/elf_linker/pkg/utils"
)

const IMAGE_BASE uint64 = 0x200000

func ResolveSymbols(ctx *LinkContext) {
	// 处理全部全局符号，为全局符号指定parent obj
	for _, file := range ctx.MobjFileList {
		file.DealGlobalSymbols()
	}

	// 标记使用到的obj文件
	// 标记规则：
	//		1.默认只有输入的obj文件是被使用状态的
	//		2.从输入的几个被使用状态的obj文件作为根节点，遍历处理这些obj文件
	//		3.遍历这些文件的全局符号，如果符号是未定义的并且所属的obj文件是未使用状态，被将其设置为使用状态
	//		4.插入到遍历数组中，形成递归处理的过程
	MarkLiveObjs(ctx)

	// 清理所有未使用到的obj文件中的符号
	for _, file := range ctx.MobjFileList {
		if !file.MisUsed {
			file.ClearSymbols()
		}
	}

	// 在全局obj文件列表中过滤掉未使用的obj文件
	ctx.MobjFileList = utils.RemoveIf(ctx.MobjFileList, func(file *InputElfObj) bool {
		return !file.MisUsed
	})
}

func MarkLiveObjs(ctx *LinkContext) {
	roots := make([]*InputElfObj, 0)
	for _, obj := range ctx.MobjFileList {
		if obj.MisUsed {
			roots = append(roots, obj)
		}
	}

	for len(roots) > 0 {

		file := roots[0]

		if !file.MisUsed {
			continue
		}

		file.MarkLiveObjs(func(f *InputElfObj) {
			roots = append(roots, f)
		})
		roots = roots[1:]
	}
}

func RegisterSectionPieces(ctx *LinkContext) {
	for _, file := range ctx.MobjFileList {
		file.RegisterSectionPieces()
	}
}

func CreateSyntheticSections(ctx *LinkContext) {
	ctx.MoutEHdr = NewElfOutputEhdr()
	ctx.Mchunks = append(ctx.Mchunks, ctx.MoutEHdr)
	ctx.MoutPHdr = NewElfOutputPhdr()
	ctx.Mchunks = append(ctx.Mchunks, ctx.MoutPHdr)
	ctx.MoutSHdr = NewElfOutputShdr()
	ctx.Mchunks = append(ctx.Mchunks, ctx.MoutSHdr)
	ctx.MgotSection = NewGotSection()
	ctx.Mchunks = append(ctx.Mchunks, ctx.MgotSection)
}

func SetOuptSectionOffsets(ctx *LinkContext) uint64 {
	addr := IMAGE_BASE
	for _, chunk := range ctx.Mchunks {
		if chunk.GetSHdr().Flags&uint64(elf.SHF_ALLOC) == 0 {
			continue
		}

		addr = utils.AlignTo(addr, chunk.GetSHdr().AddrAlign)
		chunk.GetSHdr().Addr = addr

		if !isTlsBss(chunk) {
			addr += chunk.GetSHdr().Size
		}
	}

	i := 0
	first := ctx.Mchunks[0]
	for {
		shdr := ctx.Mchunks[i].GetSHdr()
		shdr.Offset = shdr.Addr - first.GetSHdr().Addr
		i++

		if i >= len(ctx.Mchunks) || ctx.Mchunks[i].GetSHdr().Flags&uint64(elf.SHF_ALLOC) == 0 {
			break
		}
	}

	lastShdr := ctx.Mchunks[i-1].GetSHdr()
	fileoff := lastShdr.Offset + lastShdr.Size

	for ; i < len(ctx.Mchunks); i++ {
		shdr := ctx.Mchunks[i].GetSHdr()
		fileoff = utils.AlignTo(fileoff, shdr.AddrAlign)
		shdr.Offset = fileoff
		fileoff += shdr.Size
	}

	ctx.MoutPHdr.UpdateSHdr(ctx)
	return fileoff
}

func BinSections(ctx *LinkContext) {
	group := make([][]*InputElfSection, len(ctx.MoutputSections))
	for _, file := range ctx.MobjFileList {
		for _, isec := range file.MinputSections {
			if isec == nil || !isec.MisUserd {
				continue
			}

			idx := isec.MoutputSec.Midx
			group[idx] = append(group[idx], isec)
		}
	}

	for idx, osec := range ctx.MoutputSections {
		osec.Members = group[idx]
	}
}

func CollectOutputSections(ctx *LinkContext) []ElfChunker {
	osecs := make([]ElfChunker, 0)
	for _, osec := range ctx.MoutputSections {
		if len(osec.Members) > 0 {
			osecs = append(osecs, osec)
		}
	}

	for _, osec := range ctx.MmergedSections {
		if osec.Mchunk.Mhdr.Size > 0 {
			osecs = append(osecs, &osec.Mchunk)
		}
	}

	return osecs
}

func ComputeSectionSizes(ctx *LinkContext) {
	for _, osec := range ctx.MoutputSections {
		offset := uint64(0)
		p2align := int64(0)

		for _, isec := range osec.Members {
			offset = utils.AlignTo(offset, 1<<isec.Mp2Align)
			isec.Moffset = uint32(offset)
			offset += uint64(isec.MshSize)
			p2align = int64(math.Max(float64(p2align), float64(isec.Mp2Align)))
		}

		osec.Mhdr.Size = offset
		osec.Mhdr.AddrAlign = 1 << p2align
	}
}

func SortOutputSections(ctx *LinkContext) {
	getrank := func(chunk ElfChunker) int32 {
		typ := chunk.GetSHdr().Type
		flags := chunk.GetSHdr().Flags

		if flags&uint64(elf.SHF_ALLOC) == 0 {
			return math.MaxInt32 - 1
		}
		if chunk == ctx.MoutSHdr {
			return math.MaxInt32
		}
		if chunk == ctx.MoutEHdr {
			return 0
		}
		if chunk == ctx.MoutPHdr {
			return 1
		}
		if typ == uint32(elf.SHT_NOTE) {
			return 2
		}

		b2i := func(b bool) int {
			if b {
				return 1
			}
			return 0
		}

		writeable := b2i(flags&uint64(elf.SHF_WRITE) != 0)
		notExec := b2i(flags&uint64(elf.SHF_EXECINSTR) == 0)
		notTls := b2i(flags&uint64(elf.SHF_TLS) == 0)
		isBss := b2i(typ == uint32(elf.SHT_NOBITS))

		return int32(writeable<<7 | notExec<<6 | notTls<<5 | isBss<<4)
	}

	sort.SliceStable(ctx.Mchunks, func(i, j int) bool {
		return getrank(ctx.Mchunks[i]) < getrank(ctx.Mchunks[j])
	})
}

// .bss是未初始化段
func isTlsBss(chunk ElfChunker) bool {
	shdr := chunk.GetSHdr()
	return shdr.Type == uint32(elf.SHT_NOBITS) &&
		shdr.Flags&uint64(elf.SHF_TLS) != 0
}

func ComputeMergedSectionsSize(ctx *LinkContext) {
	for _, osec := range ctx.MmergedSections {
		osec.AssignOffsets()
	}
}

func ScanRelocations(ctx *LinkContext) {
	for _, file := range ctx.MobjFileList {
		file.ScanRelocations()
	}

	syms := make([]*InputElfSymbol, 0)
	for _, file := range ctx.MobjFileList {
		for _, sym := range file.MallSymbols {
			if sym.MparentFile == file && sym.Mflags != 0 {
				syms = append(syms, sym)
			}
		}
	}

	for _, sym := range syms {
		if sym.Mflags&NeedsGotTp != 0 {
			ctx.MgotSection.AddGotTpSymbol(sym)
		}

		sym.Mflags = 0
	}
}
