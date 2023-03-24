package main

import (
	"fmt"
	"os"

	"github.com/HelloYmf/elf_linker/pkg/elf"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

// Program Header Table 运行时
// Section Header Table 链接时
func main() {

	if len(os.Args) < 2 {
		utils.FatalExit("wrong args.")
	}

	file := elf.MustNewFile(os.Args[1])

	elffile := elf.LoadElf(&file.Contents)

	fmt.Println(elffile.MelfHdr)

	fmt.Println(elffile.MsectionHdr)

}
