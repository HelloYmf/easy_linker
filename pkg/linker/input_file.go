package linker

import (
	"fmt"
	"strings"

	"github.com/HelloYmf/elf_linker/pkg/file"
	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

func InputFiles(ctx *LinkContext) {
	// 处理静态链接库文件
	for _, orilibname := range (*ctx).MargsData.MstaticLibraryList {
		// 将名字恢复成.a的状态
		orilibname = strings.TrimPrefix(orilibname, "-l")
		libname := fmt.Sprintf("lib%s.a", orilibname)
		// 遍历库目录，解析archive文件
		var f *file.File = nil
		for _, libpath := range (*ctx).MargsData.MlibraryPathList {
			libpath = libpath + "/" + libname
			// 尝试打开文件
			f = file.TestNewDiskFile(libpath)
			if f != nil {
				DealFile(ctx, f)
				fmt.Printf("Load lib success: %s\n", libpath)
				break
			}
		}
		if f == nil {
			errinfo := fmt.Sprintf("Not found library: %s\n", libname)
			utils.FatalExit(errinfo)
		}
	}
	// 处理所有输入的.o文件
	for _, oriobjname := range (*ctx).MargsData.MobjPathList {
		f := file.MustNewDiskFile(oriobjname)
		fmt.Printf("Load input obj success: %s\n", f.Name)
		DealFile(ctx, f)
	}
}

func DealFile(ctx *LinkContext, f *file.File) {
	switch f.Type {
	case file.FileTypeElfObject:
		// 解析.o文件
		ef := elf_file.LoadElfObj(f)
		ctx.MargsData.MobjFileList = append(ctx.MargsData.MobjFileList, ef.Mfile)
	case file.FileTypeElfArchive:
		// 解析.a文件
		eaf := elf_file.LoadElfArchive(f)
		objs := eaf.ReadElfArchiveObjs()
		for _, obj := range *objs {
			ctx.MargsData.MobjFileList = append(ctx.MargsData.MobjFileList, obj.Mfile)
		}
	}
}
