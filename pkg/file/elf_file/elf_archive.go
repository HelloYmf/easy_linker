package elf_file

import (
	"github.com/HelloYmf/elf_linker/pkg/file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type ElfArchiveFile struct {
	Mfile file.File
}

func LoadElfArchive(f *file.File) *ElfArchiveFile {
	return &ElfArchiveFile{Mfile: *f}
}

func (ar *ElfArchiveFile) ReadElfArchiveObjs() *[]ElfObjFile {

	retobjfils := []ElfObjFile{}

	pos := 8
	strtab := []byte{}

	for len(ar.Mfile.Contents)-(pos) > 1 {
		if pos%2 == 1 {
			pos++
		}

		hdr := utils.BinRead[ElfArHeader](ar.Mfile.Contents[pos:])
		datastart := pos + int(ElfArHdrSize)
		size := hdr.hdrReadDataSize()
		pos += size + int(ElfArHdrSize)
		dataend := datastart + size
		contents := ar.Mfile.Contents[datastart:dataend]

		if hdr.hdrIsSymTab() {
			continue
		}

		// 获取字符串表，里面记录了obj文件名
		if hdr.hdrIsStrTab() {
			strtab = contents
			continue
		}

		f := LoadElfObjBuffer(contents)
		f.SetObjFileName(hdr.hdrReadName(strtab))
		f.SetObjFileParent(ar.Mfile.Name)

		retobjfils = append(retobjfils, *f)

	}
	return &retobjfils
}
