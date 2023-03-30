package linker

import (
	"math"

	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

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

func CreateInternalFile(ctx *LinkContext) {
	obj := &InputElfObj{}
	ctx.MinternalObj = obj
	ctx.MobjFileList = append(ctx.MobjFileList, obj)

	ctx.MinternalSyms = make([]elf_file.ElfSymbol, 1)
	obj.MallSymbols = append(obj.MallSymbols, NewElfInputSymbol(""))
	obj.MobjFile.MglobalSymndx = 1
	obj.MisUsed = true

	obj.MobjFile.MsymTable = ctx.MinternalSyms
}

func CreateSyntheticSections(ctx *LinkContext) {
	ctx.MoutEHdr = NewElfOutputEhdr()
	ctx.Mchunks = append(ctx.Mchunks, ctx.MoutEHdr)
	ctx.MoutSHdr = NewElfOutputShdr()
	ctx.Mchunks = append(ctx.Mchunks, ctx.MoutSHdr)
}

func SetOuptSectionOffsets(ctx *LinkContext) uint64 {
	filoff := uint64(0)

	for _, c := range ctx.Mchunks {
		filoff = utils.AlignTo(filoff, c.GetSHdr().AddrAlign)
		c.GetSHdr().Offset = filoff
		filoff += c.GetSHdr().Size
	}

	return filoff
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
