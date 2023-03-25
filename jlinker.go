package main

import (
	"fmt"
	"os"

	"github.com/HelloYmf/elf_linker/pkg/file"
	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

// Program Header Table 运行时
// Section Header Table 链接时
func main() {

	if len(os.Args) < 2 {
		utils.FatalExit("wrong args.")
	}

	file := file.MustNewFile(os.Args[1])

	objfils := elf_file.ElfObjFile{}
	objfils.LoadElfObj(&file.Contents)

	objfils.PraseSymbolTable()

	// 遍历符号表数组
	for i, sym := range objfils.MsymbolTable {
		if i >= int(objfils.Mglobalsymndx) {
			fmt.Printf("global symbol: %v\n", sym)
		} else {
			fmt.Printf("local symbol: %v\n", sym)
		}
	}
}
