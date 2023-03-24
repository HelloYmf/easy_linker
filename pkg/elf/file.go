package elf

import (
	"os"

	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type File struct {
	Name     string
	Contents []byte
}

func MustNewFile(filename string) *File {
	contents, err := os.ReadFile(filename)
	utils.MustNoErr(err)
	return &File{
		Name:     filename,
		Contents: contents,
	}
}
