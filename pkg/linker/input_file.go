package linker

import (
	"github.com/HelloYmf/elf_linker/pkg/file"
	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
)

func InputFiles(ctx *LinkContext) {
	// for libname := range (*ctx).MargsData.MstaticLibraryList {

	// }
}

func DealFile(ctx *LinkContext, f file.File) {
	switch f.Type {
	case file.FileTypeElfObject:
		ef := elf_file.LoadElfObj(f)
		ctx.MargsData.MobjFileList = append(ctx.MargsData.MobjFileList, ef.Mfile)
	case file.FileTypeElfArchive:

	}
}
