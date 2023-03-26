package linker

import "github.com/HelloYmf/elf_linker/pkg/file"

type ContextArgs struct {
	Moutput            string      // 输出目录
	March              string      // 链接器处理的架构
	MlibraryPathList   []string    // 库目录列表
	MstaticLibraryList []string    // 静态链接库文件名字列表
	MobjPathList       []string    // 输入obj文件名字列表
	MobjFileList       []file.File // obj文件对象列表
}

type LinkContext struct {
	MargsData ContextArgs
}

func NewLinkContext() LinkContext {
	return LinkContext{MargsData: ContextArgs{
		Moutput: "a.out",
		March:   "",
	}}
}
