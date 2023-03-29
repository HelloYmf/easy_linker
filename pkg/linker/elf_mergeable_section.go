package linker

import "sort"

// 可合并的节
type ElfMergeableSection struct {
	Mparent      *ElfMergedSection // 属于哪个已经合并的节
	Mp2Align     uint8             // 对齐参数
	Moridata     []string          // 原始块数据，这里设置为字符串类型为了方便比较
	MblockOffset []uint32          // 原始块在section中的偏移
	Mblock       []*ElfSectionBlock
}

func (ms *ElfMergeableSection) GetBlock(offset uint32) (*ElfSectionBlock, uint32) {
	pos := sort.Search(len(ms.MblockOffset), func(i int) bool {
		return offset < ms.MblockOffset[i]
	})

	if pos == 0 {
		return nil, 0
	}

	idx := pos - 1
	return ms.Mblock[idx], offset - ms.MblockOffset[idx]

}
