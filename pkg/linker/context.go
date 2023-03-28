package linker

type ContextArgs struct {
	Moutput            string   // 输出目录
	March              string   // 链接器处理的架构
	MlibraryPathList   []string // 库目录列表
	MstaticLibraryList []string // 静态链接库文件名字列表
	MobjPathList       []string // 输入obj文件名字列表
}

type LinkContext struct {
	MargsData    ContextArgs
	MobjFileList []*InputElfObj             // obj文件对象列表
	MsymMap      map[string]*InputElfSymbol // 整个链接过程中用到的符号
}

func NewLinkContext() LinkContext {
	return LinkContext{
		MargsData: ContextArgs{
			Moutput: "a.out",
			March:   ""},
		MsymMap: make(map[string]*InputElfSymbol),
	}
}
