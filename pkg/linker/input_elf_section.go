package linker

import (
	"debug/elf"
	"math/bits"

	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type InputElfSection struct {
	MparentFile *InputElfObj // 所属文件
	Mcontents   []byte       // 内部数据
	Mshndx      uint32       // SectionHeader数组内下标
	MshSize     uint64
	MisUserd    bool // 表示是否会被放入可执行文件中
	Mp2Align    uint8
}

func NewElfInputSection(objfil *InputElfObj, shndx uint32) *InputElfSection {
	rets := &InputElfSection{
		MparentFile: objfil,
		Mshndx:      shndx,
		MisUserd:    true,
	}

	shdr := rets.GetSectionHdr()
	rets.Mcontents = objfil.MobjFile.Mfile.Contents[shdr.Offset : shdr.Offset+shdr.Size]
	// 这里留了一个坑，压缩节的size不在这里读
	if (shdr.Flags & uint64(uint64(elf.SHF_COMPRESSED))) != 0 {
		utils.FatalExit("暂时不支持压缩节的初始化")
	}
	rets.MshSize = shdr.Size

	rets.Mp2Align = uint8(bits.TrailingZeros64(shdr.AddrAlign))

	return rets
}

func (is *InputElfSection) GetSectionHdr() *elf_file.ElfSectionHdr {
	if is.Mshndx < uint32(len(is.MparentFile.MobjFile.MsectionHdr)) {
		return &is.MparentFile.MobjFile.MsectionHdr[is.Mshndx]
	}
	return nil
}

func (is *InputElfSection) GetParentName() string {
	return is.MparentFile.MobjFile.GetSectionName(is.GetSectionHdr().Name)
}
