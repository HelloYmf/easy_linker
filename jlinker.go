package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	ctx := linker.PraseArgs(args) // 处理参数列表，初始化context
	// 优化静态库路径
	for i, path := range ctx.MargsData.MlibraryPathList {
		ctx.MargsData.MlibraryPathList[i] = filepath.Clean(path)
	}
	// 如果链接器没有给参数，就获取第一个obj文件的Machine
	if ctx.MargsData.March == "" {
		first_file := file.MustNewDiskFile(ctx.MargsData.MobjPathList[0])
		switch first_file.Type {
		case file.FileTypeElfObject: // ELF文件
			obj_file := elf_file.LoadElfObj(first_file)
			ctx.MargsData.March = obj_file.GetElfArch()
		case file.FileTypePeObject: // COFF文件
			// TODO
		}
	}

	utils.MyPrintLog("Input objs: ")
	for _, file := range ctx.MargsData.MobjPathList {
		fmt.Println("\t" + file)
	}

	utils.MyPrintLog("Library path lists: ")
	for _, path := range ctx.MargsData.MlibraryPathList {
		fmt.Println("\t" + path)
	}
	utils.MyPrintLog("Static Library lists: ")
	for _, file := range ctx.MargsData.MstaticLibraryList {
		orilibname := strings.TrimPrefix(file, "-l")
		libname := fmt.Sprintf("lib%s.a", orilibname)
		fmt.Println("\t" + libname)
	}

	// 根据处理输入的文件提取基础信息
	linker.InputFiles(&ctx)
	// 在基础信息上面处理符号之间的依赖，如解析未定义符号、删除未使用的符号和obj文件、生成唯一的全局符号map
	linker.ResolveSymbols(&ctx)
	// 将符号所属的parent更加细化（InputSection或Block）
	linker.RegisterSectionPieces(&ctx)
	linker.ComputeMergedSectionsSize(&ctx)
	// 创建输出文件
	linker.CreateSyntheticSections(&ctx)
	linker.BinSections(&ctx)
	ctx.Mchunks = append(ctx.Mchunks, linker.CollectOutputSections(&ctx)...)
	linker.ScanRelocations(&ctx)
	linker.ComputeSectionSizes(&ctx)
	linker.SortOutputSections(&ctx)

	for _, chunk := range ctx.Mchunks {
		chunk.UpdateSHdr(&ctx)
	}

	utils.MyPrintLog("Output file:")
	fmt.Println("\t" + ctx.MargsData.Moutput)

	// 输出可执行文件大小
	outfilesize := linker.SetOuptSectionOffsets(&ctx)
	ctx.Mbuf = make([]byte, outfilesize)

	utils.MyPrintLog("Output file size:")
	printsize := fmt.Sprintf("\t%v bytes", outfilesize)
	fmt.Println(printsize)

	// 创建文件
	file, err := os.OpenFile(ctx.MargsData.Moutput, os.O_RDWR|os.O_CREATE, 0777)
	utils.MustNoErr(err)

	// chunk -> buf
	for _, chunk := range ctx.Mchunks {
		chunk.CopyBuf(&ctx)
	}

	// 写入文件
	_, err = file.Write(ctx.Mbuf)
	utils.MustNoErr(err)

	utils.MyPrintLog("Output file write success.\n\n")
}
