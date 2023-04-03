package linker

import (
	"debug/elf"
	"sort"

	"github.com/HelloYmf/elf_linker/pkg/utils"
)

// 合并后的节
type ElfMergedSection struct {
	ElfChunk
	Map map[string]*ElfSectionBlock
}

func NewElfMergedSection(name string, flag uint64, typ uint64) *ElfMergedSection {
	m := ElfMergedSection{
		ElfChunk: NewThunk(),
		Map:      make(map[string]*ElfSectionBlock),
	}
	m.Mname = name
	m.Mhdr.Flags = flag
	m.Mhdr.Type = uint32(typ)

	return &m
}

func GetMergedSectionInstance(ctx *LinkContext, name string, typ uint64, flag uint64) *ElfMergedSection {
	outputname := GetOuputName(name, flag)

	// 过滤干扰位
	flag = flag & ^uint64(elf.SHF_GROUP) & ^uint64(elf.SHF_MERGE) &
		^uint64(elf.SHF_STRINGS) & ^uint64(elf.SHF_COMPRESSED)

	find := func() *ElfMergedSection {
		for _, sec := range ctx.MmergedSections {
			if outputname == sec.Mname && flag == sec.Mhdr.Flags &&
				typ == uint64(sec.Mhdr.Type) {
				return sec
			}
		}
		return nil
	}

	if sec := find(); sec != nil {
		return sec
	}

	sec := NewElfMergedSection(outputname, flag, typ)
	ctx.MmergedSections = append(ctx.MmergedSections, sec)
	return sec
}

func (ms *ElfMergedSection) Insert(key string, p2aligin uint8) *ElfSectionBlock {
	block, ok := ms.Map[key]
	if !ok {
		block = NewElfSectionBlock(ms)
		ms.Map[key] = block
	}

	if block.Mp2Align < p2aligin {
		block.Mp2Align = p2aligin
	}

	return block

}

func (ms *ElfMergedSection) AssignOffsets() {
	var blocks []struct {
		Key string
		Val *ElfSectionBlock
	}

	for key := range ms.Map {
		blocks = append(blocks, struct {
			Key string
			Val *ElfSectionBlock
		}{Key: key, Val: ms.Map[key]})
	}

	sort.SliceStable(blocks, func(i, j int) bool {
		x := blocks[i]
		y := blocks[j]
		if x.Val.Mp2Align != y.Val.Mp2Align {
			return x.Val.Mp2Align < y.Val.Mp2Align
		}
		if len(x.Key) != len(y.Key) {
			return len(x.Key) < len(y.Key)
		}

		return x.Key < y.Key
	})

	offset := uint64(0)
	p2align := uint64(0)
	for _, block := range blocks {
		offset = utils.AlignTo(offset, 1<<block.Val.Mp2Align)
		block.Val.Moffset = uint32(offset)
		offset += uint64(len(block.Key))
		if p2align < uint64(block.Val.Mp2Align) {
			p2align = uint64(block.Val.Mp2Align)
		}
	}

	ms.Mhdr.Size = utils.AlignTo(offset, 1<<p2align)
	ms.Mhdr.AddrAlign = 1 << p2align
}

func (ms *ElfMergedSection) CopyBuf(ctx *LinkContext) {
	buf := ctx.Mbuf[ms.Mhdr.Offset:]
	for key := range ms.Map {
		if block, ok := ms.Map[key]; ok {
			copy(buf[block.Moffset:], key)
		}
	}
}
