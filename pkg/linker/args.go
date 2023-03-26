package linker

import (
	"fmt"
	"os"
	"strings"

	"github.com/HelloYmf/elf_linker/pkg/utils"
)

func dashs(name string) []string {
	if len(name) == 1 {
		return []string{"-" + name}
	}
	return []string{"-" + name, "--" + name}
}

// 单标志参数
func JudgeFlags(args *[]string, name string) bool {
	for _, opt := range dashs(name) {
		if (*args)[0] == opt {
			*args = (*args)[1:]
			return true
		}
	}
	return false
}

// 后面跟参数的参数
func JudgeArgs(args *[]string, name string) (bool, string) {
	arg := ""
	for _, opt := range dashs(name) {
		// 支持常规linker -o [path]参数
		if (*args)[0] == opt {
			if len(*args) == 1 {
				utils.FatalExit(fmt.Sprintf("option -%s: argument missing\n", name))
			}
			arg = (*args)[1]
			*args = (*args)[2:]
			return true, arg
		}
		prefix := opt
		// 支持linker -plugin-opt=/usr/lib前缀参数
		if len(name) > 1 {
			prefix += "="
		}
		// 支持linker -melf32前缀参数
		if strings.HasPrefix((*args)[0], prefix) {
			arg = (*args)[0][len(prefix):]
			*args = (*args)[1:]
			return true, arg
		}
	}
	return false, arg
}

func PraseArgs(args []string) LinkContext {
	ctx := NewLinkContext()
	for len(args) > 0 {
		isArg := false
		curArg := ""

		if JudgeFlags(&args, "help") {
			// 这里可以判断一下都存了哪些参数，在进一步提供帮助信息
			fmt.Printf("usage: %s [options] files...\n", os.Args[0])
			os.Exit(0)
		}

		// 解析输出目录
		isArg, curArg = JudgeArgs(&args, "o")
		if isArg {
			ctx.MargsData.Moutput = curArg
			continue
		}
		// 解析架构
		isArg, curArg = JudgeArgs(&args, "m")
		if isArg {
			if curArg == "elf64lriscv" {
				ctx.MargsData.March = "elf64"
			} else {
				utils.FatalExit(fmt.Sprintf("option -m: unknown arch %s.\n", curArg))
			}
			continue
		}
		// 解析库目录
		isArg, curArg = JudgeArgs(&args, "L")
		if isArg {
			ctx.MargsData.MlibraryPathList = append(ctx.MargsData.MlibraryPathList, curArg)
			continue
		}
		// 处理l参数（保留）
		isArg, curArg = JudgeArgs(&args, "l")
		if isArg {
			ctx.MargsData.MstaticLibraryList = append(ctx.MargsData.MstaticLibraryList, "-l"+curArg)
			continue
		}
		// 忽略列表
		isArg, curArg = JudgeArgs(&args, "sysroot")
		if isArg {
			fmt.Printf("Ignore arg: sysroot %s\n", curArg)
			continue
		}
		isArg = JudgeFlags(&args, "static")
		if isArg {
			fmt.Printf("Ignore flag: static\n")
			continue
		}
		isArg, curArg = JudgeArgs(&args, "plugin")
		if isArg {
			fmt.Printf("Ignore arg: plugin %s\n", curArg)
			continue
		}
		isArg, curArg = JudgeArgs(&args, "plugin-opt")
		if isArg {
			fmt.Printf("Ignore arg: plugin-opt %s\n", curArg)
			continue
		}
		isArg = JudgeFlags(&args, "as-needed")
		if isArg {
			fmt.Printf("Ignore flag: as-needed\n")
			continue
		}
		isArg = JudgeFlags(&args, "start-group")
		if isArg {
			fmt.Printf("Ignore flag: start-group\n")
			continue
		}
		isArg = JudgeFlags(&args, "end-group")
		if isArg {
			fmt.Printf("Ignore flag: end-group\n")
			continue
		}
		isArg, curArg = JudgeArgs(&args, "hash-style")
		if isArg {
			fmt.Printf("Ignore arg: hash-style %s\n", curArg)
			continue
		}
		isArg, curArg = JudgeArgs(&args, "build-id")
		if isArg {
			fmt.Printf("Ignore arg: build-id %s\n", curArg)
			continue
		}
		isArg = JudgeFlags(&args, "s")
		if isArg {
			fmt.Printf("Ignore flag: s\n")
			continue
		}
		isArg = JudgeFlags(&args, "no-relax")
		if isArg {
			fmt.Printf("Ignore flag: no-relax\n")
			continue
		}

		if args[0][0] == '-' {
			utils.FatalExit(fmt.Sprintf("Wrong option %s.\n", args[0]))
		}

		ctx.MargsData.MobjPathList = append(ctx.MargsData.MobjPathList, args[0])
		args = args[1:]
	}
	return ctx
}
