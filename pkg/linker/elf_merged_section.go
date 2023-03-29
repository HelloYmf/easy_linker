package linker

import "debug/elf"

// 合并后的节
type ElfMergedSection struct {
	Mchunk ELfThunk
	Map    map[string]*ElfSectionBlock
}

func NewElfMergedSection(name string, flag uint64, typ uint64) *ElfMergedSection {
	m := ElfMergedSection{
		Mchunk: NewThunk(),
		Map:    make(map[string]*ElfSectionBlock),
	}
	m.Mchunk.Mname = name
	m.Mchunk.Mhdr.Flags = flag
	m.Mchunk.Mhdr.Type = uint32(typ)

	return &m
}

func GetMergedSectionInstance(ctx *LinkContext, name string, typ uint64, flag uint64) *ElfMergedSection {
	outputname := GetOuputName(name, flag)

	// 过滤干扰位
	flag = flag & ^uint64(elf.SHF_GROUP) & ^uint64(elf.SHF_MERGE) &
		^uint64(elf.SHF_STRINGS) & ^uint64(elf.SHF_COMPRESSED)

	find := func() *ElfMergedSection {
		for _, sec := range ctx.MmergedSections {
			if outputname == sec.Mchunk.Mname && flag == sec.Mchunk.Mhdr.Flags &&
				typ == uint64(sec.Mchunk.Mhdr.Type) {
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
