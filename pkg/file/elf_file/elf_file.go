package elf_file

import (
	"bytes"

	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type ElfFile struct {
	Mcontents   []byte       // FILE_BUFFER
	MelfHdr     ElfHdr       // ELF_HEADER
	MsectionHdr []SectionHdr // ELF_SECTION_HEADER[]
}

func (f *ElfFile) LoadElf(contents *[]byte) {

	f.Mcontents = *contents

	// 判断是否是有效的ELF文件
	if !bytes.HasPrefix(*contents, []byte("\x7fELF")) {
		utils.FatalExit("Invalid ELF file.")
	}

	// 判断ELF文件头部是否标准
	if len(*contents) < int(ElfHdrSize) {
		utils.FatalExit("ELF header is too short.")
	}

	// 读取ElfHeader数据
	elfHdr := utils.BinRead[ElfHdr](*contents)
	f.MelfHdr = elfHdr

	// 读取第一个SectionHeader数据
	context := (*contents)[elfHdr.ShOff:]
	sHdr := utils.BinRead[SectionHdr](context)

	// 如果Section的数量很多，超出了elfHdr.ShNum字段uint16的限制，此时真正的大小在第一个SectionHdr中的Size字段uint64
	numSections := uint64(elfHdr.ShNum)
	if numSections == 0 {
		numSections = sHdr.Size
	}
	f.MsectionHdr = []SectionHdr{sHdr}
	// 循环读取SectionHeader数据
	for numSections > 1 {
		context = context[SectionHdrSize:]
		f.MsectionHdr = append(f.MsectionHdr, utils.BinRead[SectionHdr](context))
		numSections--
	}
}

func (f *ElfFile) GetSectionData(secndx int64) (secdata []byte) {
	hdr := f.MsectionHdr[secndx]
	return f.Mcontents[hdr.Offset : hdr.Offset+hdr.Size]
}

func (f *ElfFile) GetTargetTypeOfSections(sectype uint32) (ndxarr []int) {
	ret := []int{}
	for i, hdr := range f.MsectionHdr {
		if hdr.Type == sectype {
			ret = append(ret, i)
		}
	}
	return ret
}
