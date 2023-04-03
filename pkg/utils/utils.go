package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/bits"
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

func MyPrintLog(v any) {
	fmt.Printf("\033[0;0;32m\n\t\t%v\033[0m \n", v)
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

func RemoveIf[T any](elems []T, condition func(T) bool) []T {
	i := 0
	for _, elem := range elems {
		if condition(elem) {
			continue
		}
		elems[i] = elem
		i++
	}
	return elems[:i]
}

func AllZeros(bs []byte) bool {
	b := byte(0)
	for _, s := range bs {
		b |= s
	}

	return b == 0
}

func AlignTo(val, align uint64) uint64 {
	if align == 0 {
		return val
	}

	return (val + align - 1) &^ (align - 1)
}

func Write[T any](data []byte, e T) {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, e)
	MustNoErr(err)
	copy(data, buf.Bytes())
}

func ReadSlice[T any](data []byte, sz int) []T {
	nums := len(data) / sz
	res := make([]T, 0, nums)
	for nums > 0 {
		res = append(res, BinRead[T](data))
		data = data[sz:]
		nums--
	}

	return res
}

func hasSingleBit(n uint64) bool {
	return n&(n-1) == 0
}

func BitCeil(val uint64) uint64 {
	if hasSingleBit(val) {
		return val
	}
	return 1 << (64 - bits.LeadingZeros64(val))
}

type Uint interface {
	uint8 | uint16 | uint32 | uint64
}

func Bit[T Uint](val T, pos int) T {
	return (val >> pos) & 1
}

func Bits[T Uint](val T, hi T, lo T) T {
	return (val >> lo) & ((1 << (hi - lo + 1)) - 1)
}

func SignExtend(val uint64, size int) uint64 {
	return uint64(int64(val<<(63-size)) >> (63 - size))
}
