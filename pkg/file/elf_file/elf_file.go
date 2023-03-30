package elf_file

import (
	"bytes"
	"debug/elf"

	"github.com/HelloYmf/elf_linker/pkg/file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type ElfFile struct {
	Mfile            file.File       // file基类
	MelfHdr          ElfHdr          // ELF_HEADER
	MsectionHdr      []ElfSectionHdr // ELF_SECTION_HEADER[]
	MsectionNameData []byte          // 存储节区名字的section内容
}

func LoadElfBuffer(contents []byte) *ElfFile {
	f := file.NewMemoryFile(contents)
	return InitElf(f)
}

func LoadElfFile(filename string) *ElfFile {
	f := file.MustNewDiskFile(filename)
	return InitElf(f)
}

func LoadElf(f *file.File) *ElfFile {
	return InitElf(f)
}

func InitElf(f *file.File) *ElfFile {
	ret := new(ElfFile)
	(*ret).Mfile = *f
	// 判断是否是有效的ELF文件
	if f.Type != file.FileTypeElfExe && f.Type != file.FileTypeElfObject && f.Type != file.FileTypeElfSo {
		utils.FatalExit("Invalid ELF file.")
	}

	// 判断ELF文件头部是否标准
	if len(f.Contents) < int(ElfHdrSize) {
		utils.FatalExit("ELF header is too short.")
	}

	// 读取ElfHeader数据
	elfHdr := utils.BinRead[ElfHdr](f.Contents)
	(*ret).MelfHdr = elfHdr

	// 读取第一个SectionHeader数据
	context := (f.Contents)[elfHdr.ShOff:]
	sHdr := utils.BinRead[ElfSectionHdr](context)

	// 如果Section的数量很多，超出了elfHdr.ShNum字段uint16的限制，此时真正的大小在第一个SectionHdr中的Size字段uint64
	numSections := uint64(elfHdr.ShNum)
	if numSections == 0 {
		numSections = sHdr.Size
	}
	(*ret).MsectionHdr = []ElfSectionHdr{sHdr}
	// 循环读取SectionHeader数据
	for numSections > 1 {
		context = context[ElfSectionHdrSize:]
		ret.MsectionHdr = append((*ret).MsectionHdr, utils.BinRead[ElfSectionHdr](context))
		numSections--
	}
	return ret
}

func (f *ElfFile) GetSectionData(secndx int64) (secdata []byte) {
	hdr := f.MsectionHdr[secndx]
	return f.Mfile.Contents[hdr.Offset : hdr.Offset+hdr.Size]
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

func (f *ElfFile) GetElfArch() string {
	// 判断处理器架构
	switch f.MelfHdr.Machine {
	case uint16(elf.EM_RISCV):
		return "elf64lriscv"
	}
	return "unknown"
}

func (f *ElfFile) GetSectionName(name_index uint32) string {
	// 避免重复获取
	if len(f.MsectionNameData) == 0 {
		f.getSectionNameData()
	}
	if len(f.MsectionNameData) == 0 {
		return ""
	}
	// 获取长度转成字符串
	namelength := uint32(bytes.Index(f.MsectionNameData[name_index:], []byte{0}))
	return string(f.MsectionNameData[name_index : name_index+namelength])
}

func (f *ElfFile) getSectionNameData() {
	// 获取节名字所在的section的索引
	shstrndx := int64(f.MelfHdr.ShStrndx)
	// 如果ELF_HEADER中字段ShStrndx的2字节存不下
	if f.MelfHdr.ShStrndx == uint16(elf.SHN_XINDEX) {
		// 此时真正的索引在第一个SectionHeader的Link字段3字节
		shstrndx = int64(f.MsectionHdr[0].Link)
	}
	// 获取存放SectionName字符串的Section数据
	sectionname := f.GetSectionData(shstrndx)
	f.MsectionNameData = sectionname
}
