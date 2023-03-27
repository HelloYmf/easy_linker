package main

import (
	"os"

	"github.com/HelloYmf/elf_linker/pkg/file"
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

	f := file.NewDiskFile(os.Args[2])
	linker.DealFile(&ctx, f)

	// // 如果链接器没有给参数，就获取第一个obj文件的Machine
	// if ctx.MargsData.March == "" {
	// 	first_file := file.NewDiskFile(ctx.MargsData.MobjPathList[0])
	// 	switch first_file.Type {
	// 	case file.FileTypeElfObject:
	// 		obj_file := elf_file.LoadElfObj(first_file)
	// 		ctx.MargsData.March = obj_file.GetElfArch()
	// 	case file.FileTypePeObject:
	// 		// TODO
	// 	}
	// }

	// fmt.Printf("Output path: %s\n", ctx.MargsData.Moutput)
	// fmt.Printf("Arch: %s\n", ctx.MargsData.March)
	// fmt.Printf("MlibraryPath: %v\n", ctx.MargsData.MlibraryPathList)
	// fmt.Printf("MobjPathList: %v\n", ctx.MargsData.MobjPathList)
	// fmt.Printf("StaticLibraryList: %v\n", ctx.MargsData.MstaticLibraryList)

	// objfils := elf_file.LoadElfObjFile(ctx.MargsData.MobjPathList[0])
	// objfils.PraseSymbolTable()
	// // 遍历符号表数组
	// for _, sym := range objfils.MsymTable {
	// 	symname := objfils.GetSymbolName(sym.Name)
	// 	if len(symname) != 0 {
	// 		fmt.Printf("\t%s\n", symname)
	// 	}
	// }
}
