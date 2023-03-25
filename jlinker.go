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

	// 如果链接器没有给参数，就获取第一个obj文件的类型
	if ctx.MargsData.March == "" {
		first_file := file.MustNewFile(os.Args[1])
		obj_file := elf_file.LoadElfObj(&first_file.Contents)
		ctx.MargsData.March = obj_file.GetElfArch()
	}

	fmt.Printf("Output path: %s\n", ctx.MargsData.Moutput)
	fmt.Printf("Arch: %s\n", ctx.MargsData.March)
	fmt.Printf("MlibraryPath: %v\n", ctx.MargsData.MlibraryPath)
	fmt.Printf("MobjList: %v\n", ctx.MargsData.MobjList)
	fmt.Printf("Remaining: %v\n", ctx.MargsData.Mremaining)

	//		file := file.MustNewFile(os.Args[1])
	//		objfils := elf_file.LoadElfObj(&file.Contents)
	//		objfils.PraseSymbolTable()
	//		// 遍历符号表数组
	//		for _, sym := range objfils.MsymTable {
	//			symname := objfils.GetSymbolName(sym.Name)
	//			if len(symname) != 0 {
	//				fmt.Printf("\t%s\n", symname)
	//			}
	//		}
	//	}
}
