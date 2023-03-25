package elf_file

import (
	"bytes"
	"debug/elf"

	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type ElfObjFile struct {
	ElfFile
	MsectionNameData []byte
	MsymTable        []Symbol
	MglobalSymndx    uint32
	MsymNameData     []byte
}

func (f *ElfObjFile) LoadElfObj(contents *[]byte) {
	f.LoadElf(contents)
}

func (f *ElfObjFile) getSectionNameData() {
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

func (f *ElfObjFile) GetSectionName(name_index uint32) string {
	// 避免重复获取
	if len(f.MsectionNameData) == 0 {
		f.getSectionNameData()
	}
	if len(f.MsectionNameData) != 0 {
		// 获取长度转成字符串
		namelength := uint32(bytes.Index(f.MsectionNameData[name_index:], []byte{0}))
		return string(f.MsectionNameData[name_index : name_index+namelength])
	}
	return ""
}

func (f *ElfObjFile) PraseSymbolTable() {
	symndx := f.GetTargetTypeOfSections(uint32(elf.SHT_SYMTAB))[0]
	symdata := f.GetSectionData(int64(symndx))
	symnum := int64(len(symdata)) / int64(SymbolSize)
	// 获取符号名字section数据
	symname := f.GetSectionData(int64(f.MsectionHdr[symndx].Link))
	f.MsymNameData = symname

	// 解析符号表数组
	f.MsymTable = []Symbol{utils.BinRead[Symbol](symdata)}
	for i := int64(1); i < symnum; i++ {
		symdata = symdata[SymbolSize:]
		f.MsymTable = append(f.MsymTable, utils.BinRead[Symbol](symdata))
	}

	// 因为符号分为Local符号和Global符号，而ELF中Local在前Global在后，SymbolHader的Info就是Global符号在符号表中的起始位置
	f.MglobalSymndx = f.MsectionHdr[symndx].Info
}

func (f *ElfObjFile) GetSymbolName(name_index uint32) string {
	// 避免重复获取
	if len(f.MsymNameData) == 0 {
		f.PraseSymbolTable()
	}
	if len(f.MsectionNameData) != 0 {
		// 获取长度转成字符串
		namelength := uint32(bytes.Index(f.MsymNameData[name_index:], []byte{0}))
		return string(f.MsymNameData[name_index : name_index+namelength])
	}
	return ""
}