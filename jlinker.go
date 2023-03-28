package main

import (
	"fmt"
	"os"

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

	fmt.Printf("total loaded objs: %d\n", len(ctx.MobjFileList))

	for _, obj := range ctx.MobjFileList {
		if obj.MobjFile.Mfile.Name == "out/tests/call_api/call_api.o" {
			for _, sym := range obj.MallSymbols {
				if sym.Mname != "" {
					fmt.Printf("sym: %s from: %s\n", sym.Mname, sym.MparentFile.MobjFile.Mfile.Name)
				}
			}
		}
	}
}
