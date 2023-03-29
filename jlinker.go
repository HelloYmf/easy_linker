package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HelloYmf/elf_linker/pkg/file"
	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
	"github.com/HelloYmf/elf_linker/pkg/linker"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

// Program Header Table 运行时
// Section Header Table 链接时
func main() {
	if len(os.Args) < 2 {
		utils.FatalExit("wrong args.")
	}

	args := os.Args[1:]
	ctx := linker.PraseArgs(args)

	// 优化路径
	for i, path := range ctx.MargsData.MlibraryPathList {
		ctx.MargsData.MlibraryPathList[i] = filepath.Clean(path)
	}

	// 如果链接器没有给参数，就获取第一个obj文件的Machine
	if ctx.MargsData.March == "" {
		first_file := file.MustNewDiskFile(ctx.MargsData.MobjPathList[0])
		switch first_file.Type {
		case file.FileTypeElfObject:
			obj_file := elf_file.LoadElfObj(first_file)
			ctx.MargsData.March = obj_file.GetElfArch()
		case file.FileTypePeObject:
			// TODO
		}
	}

	linker.InputFiles(&ctx)
	linker.ResolveSymbols(&ctx)
	linker.RegisterSectionPieces(&ctx)

	fmt.Printf("total loaded objs: %d\n", len(ctx.MobjFileList))
	fmt.Printf("total merged sectins: %d\n", len(ctx.MmergedSections))

	for _, sec := range ctx.MmergedSections {
		fmt.Printf("%v\n", sec.Mchunk.Mname)
	}

}
