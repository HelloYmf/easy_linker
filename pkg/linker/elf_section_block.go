package linker

import "math"

type ElfSectionBlock struct {
	MmergedSection *ElfMergedSection // 所属的合并节
	Moffset        uint32            // 在section中的偏移
	Mp2Align       uint8             // 对齐参数
	MisUserd       bool              // 标记哪些可以合并，保留一份即可
}

func NewElfSectionBlock(section *ElfMergedSection) *ElfSectionBlock {
	return &ElfSectionBlock{Moffset: math.MaxUint32, MmergedSection: section}
}

func (sb *ElfSectionBlock) GetAddr() uint64 {
	return sb.MmergedSection.Mchunk.Mhdr.Addr + uint64(sb.Moffset)
}
