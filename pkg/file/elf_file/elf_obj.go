package elf_file

import (
	"bytes"
	"debug/elf"

	"github.com/HelloYmf/elf_linker/pkg/file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type ElfObjFile struct {
	ElfFile                   // elf基类
	MsymTable     []ElfSymbol // 符号项数组
	MglobalSymndx uint32      // 全局符号在符号数组中的起始偏移
	MsymNameData  []byte      // 存储符号名字的section数据
	MlibName      string      // 所属lib
}

func LoadElfObjBuffer(contents []byte) *ElfObjFile {
	return &ElfObjFile{ElfFile: *LoadElfBuffer(contents)}
}

func LoadElfObjFile(filename string) *ElfObjFile {
	return &ElfObjFile{ElfFile: *LoadElfFile(filename)}
}

func LoadElfObj(f *file.File) *ElfObjFile {
	return &ElfObjFile{ElfFile: *LoadElf(f)}
}

func (f *ElfObjFile) SetObjFileName(name string) {
	f.ElfFile.Mfile.Name = name
}

func (f *ElfObjFile) SetObjFileParent(parent string) {
	f.MlibName = parent
}

func (f *ElfObjFile) PraseSymbolTable() {
	// 根据section hdr中的Type字段获取符号section在section hdr数组中的索引
	// 返回一个数组，因为同种Type的section hdr可能不止一个
	secidx := f.GetTargetTypeOfSections(uint32(elf.SHT_SYMTAB))
	if len(secidx) == 0 {
		return
	}
	symndx := secidx[0]
	// 获取符号表数据
	symdata := f.GetSectionData(int64(symndx))
	// 符号表是一个ElfSymbol结构的数组，每个符号项对应一个ElfSymbol结构
	symnum := int64(len(symdata)) / int64(ElfSymbolSize)
	// 获取符号名字section数据，符号header的Link字段定义了符号名字所在的hdr索引
	symnamedata := f.GetSectionData(int64(f.MsectionHdr[symndx].Link))
	f.MsymNameData = symnamedata

	// 解析符号项数组
	f.MsymTable = []ElfSymbol{utils.BinRead[ElfSymbol](symdata)}
	for i := int64(1); i < symnum; i++ {
		symdata = symdata[ElfSymbolSize:]
		f.MsymTable = append(f.MsymTable, utils.BinRead[ElfSymbol](symdata))
	}

	// 因为符号分为Local符号和Global符号，而ELF中Local在前Global在后，SymbolHader的Info就是Global符号在符号表中的起始位置
	f.MglobalSymndx = f.MsectionHdr[symndx].Info
}

func (f *ElfObjFile) GetSymbolName(name_index uint32) string {
	// 避免重复获取
	if len(f.MsymNameData) == 0 {
		f.PraseSymbolTable()
	}
	if len(f.MsymNameData) == 0 {
		return ""
	}
	// 获取长度转成字符串
	namelength := uint32(bytes.Index(f.MsymNameData[name_index:], []byte{0}))
	return string(f.MsymNameData[name_index : name_index+namelength])
}
