package linker

type ContextArgs struct {
	Moutput      string
	March        string
	MlibraryPath []string
	MobjList     []string
	Mremaining   []string
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
