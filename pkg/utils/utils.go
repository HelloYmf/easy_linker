package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"runtime/debug"
)

func FatalExit(v any) {
	fmt.Printf("\033[0;1;31mfatal: %v\033[0m \n", v)
	fmt.Printf("\n\033[0;1;35m")
	debug.PrintStack()
	fmt.Printf("\033[0m\n")
	os.Exit(-1)
}

func MustNoErr(err error) {
	if err != nil {
		FatalExit(err)
	}
}

func BinRead[T any](data []byte) (val T) {
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, &val)
	MustNoErr(err)
	return val
}
