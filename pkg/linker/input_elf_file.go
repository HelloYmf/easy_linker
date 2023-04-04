package linker

import (
	"fmt"
	"strings"

	"github.com/HelloYmf/elf_linker/pkg/file"
	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

func InputFiles(ctx *LinkContext) {
	// 处理所有输入的.o文件
	for _, oriobjname := range (*ctx).MargsData.MobjPathList {
		f := file.MustNewDiskFile(oriobjname)
		DealFile(ctx, f)
	}
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
				break
			}
		}
		if f == nil {
			errinfo := fmt.Sprintf("Not found library: %s\n", libname)
			utils.FatalExit(errinfo)
		}
	}
}

func DealFile(ctx *LinkContext, f *file.File) {
	switch f.Type {
	case file.FileTypeElfObject:
		// 加载elf obj文件
		// elf_obj是elf的子类，elf中会初始化elf header、section header[]，提供获取section name的方法
		// elf_obj中提供了
		ef := elf_file.LoadElfObj(f)
		// 每个obj文件检查一次架构
		CheckMachine(ctx, ef)
		// 使用elf文件加载Input obj
		// InputFile是Elf
		inputfil := NewElfInputObj(ctx, ef)
		inputfil.MisUsed = true

		ctx.MobjFileList = append(ctx.MobjFileList, inputfil)
	case file.FileTypeElfArchive:
		// 解析.a文件
		eaf := elf_file.LoadElfArchive(f)
		objs := eaf.ReadElfArchiveObjs()
		for _, obj := range *objs {
			CheckMachine(ctx, &obj)
			inputfil := NewElfInputObj(ctx, &obj)

			inputfil.MisUsed = false
			ctx.MobjFileList = append(ctx.MobjFileList, inputfil)
		}
	}
}

// 检查obj文件架构一致性
func CheckMachine(ctx *LinkContext, f *elf_file.ElfObjFile) {
	if f.GetElfArch() != ctx.MargsData.March {
		utils.FatalExit("Inconsistent architecture")
	}
}
