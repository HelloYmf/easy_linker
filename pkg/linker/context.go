package linker

import "github.com/HelloYmf/elf_linker/pkg/file/elf_file"

type ContextArgs struct {
	Moutput            string   // 输出目录
	March              string   // 链接器处理的架构
	MlibraryPathList   []string // 库目录列表
	MstaticLibraryList []string // 静态链接库文件名字列表
	MobjPathList       []string // 输入obj文件名字列表
}

type LinkContext struct {
	MargsData       ContextArgs
	MobjFileList    []*InputElfObj             // obj文件对象列表
	MsymMap         map[string]*InputElfSymbol // 整个链接过程中用到的符号
	MmergedSections []*ElfMergedSection        // 所有要输出的合并后的section
	MinternalObj    *InputElfObj
	MinternalSyms   []elf_file.ElfSymbol

	Mchunks         []ElfChunker // 所有要写入可执行文件的元素的基类
	MoutputSections []*ElfOutputSection
	Mbuf            []byte

	MoutEHdr *ElfOutputEhdr // 输出可执行文件中的ELF Header
	MoutSHdr *ElfOutputShdr // 输出可执行文件中的Section Header
}

func NewLinkContext() LinkContext {
	return LinkContext{
		MargsData: ContextArgs{
			Moutput: "a.out",
			March:   ""},
		MsymMap: make(map[string]*InputElfSymbol),
	}
}
