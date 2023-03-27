package linker

import (
	"fmt"

	"github.com/HelloYmf/elf_linker/pkg/file"
	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
)

func InputFiles(ctx *LinkContext) {
	// for libname := range (*ctx).MargsData.MstaticLibraryList {

	// }
}

func DealFile(ctx *LinkContext, f *file.File) {
	switch f.Type {
	case file.FileTypeElfObject:
		ef := elf_file.LoadElfObj(f)
		ctx.MargsData.MobjFileList = append(ctx.MargsData.MobjFileList, ef.Mfile)
	case file.FileTypeElfArchive:
		// 解析.a文件
		eaf := elf_file.LoadElfArchive(f)
		objs := eaf.ReadElfArchiveObjs()
		for _, obj := range *objs {
			fmt.Printf("%s\r\n", obj.Mfile.Name)
		}
	}
}
