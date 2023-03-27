package file

import (
	"bytes"
	"debug/elf"
	"os"

	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type FileType = uint8

const (
	FileTypeUnknown FileType = iota
	FileTypeEmpty   FileType = iota
	// Linux
	FileTypeElfExe     FileType = iota
	FileTypeElfObject  FileType = iota
	FileTypeElfArchive FileType = iota
	FileTypeElfSo      FileType = iota
	// Windows
	FileTypePeExe    FileType = iota
	FileTypePeObject FileType = iota
	FileTypePeLib    FileType = iota
	FileTypePeDll    FileType = iota
)

type File struct {
	Name     string
	Contents []byte
	Type     FileType
}

func MustNewDiskFile(filename string) *File {
	contents, err := os.ReadFile(filename)
	utils.MustNoErr(err)
	ft := judgeFileType(contents)
	return &File{
		Name:     filename,
		Contents: contents,
		Type:     ft,
	}
}

func TestNewDiskFile(filename string) *File {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}
	ft := judgeFileType(contents)
	return &File{
		Name:     filename,
		Contents: contents,
		Type:     ft,
	}
}

func NewMemoryFile(contents []byte) *File {
	ft := judgeFileType(contents)
	return &File{
		Name:     "",
		Contents: contents,
		Type:     ft,
	}
}

func judgeFileType(contents []byte) FileType {

	if len(contents) == 0 {
		return FileTypeEmpty
	}

	// 判断是否是ELF文件
	if isELF(contents) {
		et := elf.Type(utils.BinRead[uint16](contents[16:]))
		switch et {
		case elf.ET_EXEC:
			// .out
			return FileTypeElfExe
		case elf.ET_REL:
			// .o
			return FileTypeElfObject
		case elf.ET_DYN:
			// .so
			return FileTypeElfSo
		}
	}
	// 判断是否是Linux下的静态链接库文件（.a）
	if bytes.HasPrefix(contents, []byte("!<arch>\n")) {
		return FileTypeElfArchive
	}

	return FileTypeUnknown
}

func isELF(contents []byte) bool {
	return bytes.HasPrefix(contents, []byte("\x7fELF"))
}
