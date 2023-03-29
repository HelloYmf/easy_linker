package linker

import (
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

func ResolveSymbols(ctx *LinkContext) {
	for _, file := range ctx.MobjFileList {
		file.ResolveSymbols()
	}

	MarkLiveObjs(ctx)

	for _, file := range ctx.MobjFileList {
		if !file.MisUsed {
			file.ClearSymbols()
		}
	}

	ctx.MobjFileList = utils.RemoveIf(ctx.MobjFileList, func(file *InputElfObj) bool {
		return !file.MisUsed
	})
}

func MarkLiveObjs(ctx *LinkContext) {
	roots := make([]*InputElfObj, 0)
	for _, obj := range ctx.MobjFileList {
		if obj.MisUsed {
			roots = append(roots, obj)
		}
	}

	for len(roots) > 0 {

		file := roots[0]

		if !file.MisUsed {
			continue
		}

		file.MarkLiveObjs(ctx, func(f *InputElfObj) {
			roots = append(roots, f)
		})
		roots = roots[1:]
	}
}

func RegisterSectionPieces(ctx *LinkContext) {
	for _, file := range ctx.MobjFileList {
		file.RegisterSectionPieces()
	}
}
