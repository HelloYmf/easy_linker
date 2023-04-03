package linker

import (
	"debug/elf"
	"math"
	"math/bits"

	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type InputElfSection struct {
	MparentFile *InputElfObj // 所属文件
	Mcontents   []byte       // 内部数据
	Mshndx      uint32       // SectionHeader数组内下标
	MshSize     uint64
	MisUserd    bool // 表示是否会被放入可执行文件中
	Mp2Align    uint8
	Moffset     uint32            // 在Output section中的偏移
	MoutputSec  *ElfOutputSection // 属于哪个output section

	MrelSecIdx uint32
	Mrels      []elf_file.ElfRela // 重定位块的数组
}

func NewElfInputSection(ctx *LinkContext, name string, objfil *InputElfObj, shndx uint32) *InputElfSection {
	rets := &InputElfSection{
		MparentFile: objfil,
		Mshndx:      shndx,
		MisUserd:    true,
		Moffset:     math.MaxUint32,
		MrelSecIdx:  math.MaxUint32,
		MshSize:     math.MaxUint32,
	}

	shdr := rets.GetSectionHdr()
	rets.Mcontents = objfil.MobjFile.Mfile.Contents[shdr.Offset : shdr.Offset+shdr.Size]
	// 这里留了一个坑，压缩节的size不在这里读
	if (shdr.Flags & uint64(uint64(elf.SHF_COMPRESSED))) != 0 {
		utils.FatalExit("暂时不支持压缩节的初始化")
	}
	rets.MshSize = shdr.Size
	rets.Mp2Align = uint8(bits.TrailingZeros64(shdr.AddrAlign))

	rets.MoutputSec = GetOutputSection(ctx, name, uint64(shdr.Type), shdr.Flags)

	return rets
}

func (is *InputElfSection) GetSectionHdr() *elf_file.ElfSectionHdr {
	if is.Mshndx < uint32(len(is.MparentFile.MobjFile.MsectionHdr)) {
		return &is.MparentFile.MobjFile.MsectionHdr[is.Mshndx]
	}
	return nil
}

func (is *InputElfSection) GetSectionName() string {
	return is.MparentFile.MobjFile.GetSectionName(is.GetSectionHdr().Name)
}

func (is *InputElfSection) WriteToBuf(ctx *LinkContext, buf []byte) {

	if is.GetSectionHdr().Type == uint32(elf.SHT_NOBITS) || is.MshSize == 0 {
		return
	}

	is.CopyContents(buf)

	if is.GetSectionHdr().Flags&uint64(elf.SHF_ALLOC) != 0 {
		is.ApplyRelocAlloc(ctx, buf)
	}
}

func (is *InputElfSection) CopyContents(buf []byte) {
	copy(buf, is.Mcontents)
}

func (is *InputElfSection) GetRels() []elf_file.ElfRela {
	if is.MrelSecIdx == math.MaxUint32 || is.Mrels != nil {
		return is.Mrels
	}

	bs := is.MparentFile.GetBytesFromShdr(&is.MparentFile.MobjFile.MsectionHdr[is.MrelSecIdx])
	is.Mrels = utils.ReadSlice[elf_file.ElfRela](bs, int(elf_file.ElfRelaSize))
	return is.Mrels
}

func (is *InputElfSection) GetAddr() uint64 {
	return is.MoutputSec.Mhdr.Addr + uint64(is.Moffset)
}

func (is *InputElfSection) ScanAllRelocations() {
	for _, rel := range is.GetRels() {
		sym := is.MparentFile.MallSymbols[rel.Sym]
		if sym.MparentFile == nil {
			continue
		}

		if rel.Type == uint32(elf.R_RISCV_TLS_GOT_HI20) {
			sym.Mflags |= NeedsGotTp
		}
	}
}

func (i *InputElfSection) ApplyRelocAlloc(ctx *LinkContext, base []byte) {
	rels := i.GetRels()

	for a := 0; a < len(rels); a++ {
		rel := rels[a]
		if rel.Type == uint32(elf.R_RISCV_NONE) ||
			rel.Type == uint32(elf.R_RISCV_RELAX) {
			continue
		}

		sym := i.MparentFile.MallSymbols[rel.Sym]
		loc := base[rel.Offset:]

		if sym.MparentFile == nil {
			continue
		}

		S := sym.GetAddr()
		A := uint64(rel.Addend)
		P := i.GetAddr() + rel.Offset

		switch elf.R_RISCV(rel.Type) {
		case elf.R_RISCV_32:
			utils.Write(loc, uint32(S+A))
		case elf.R_RISCV_64:
			utils.Write(loc, S+A)
		case elf.R_RISCV_BRANCH:
			writeBtype(loc, uint32(S+A-P))
		case elf.R_RISCV_JAL:
			writeJtype(loc, uint32(S+A-P))
		case elf.R_RISCV_CALL, elf.R_RISCV_CALL_PLT:
			val := uint32(S + A - P)
			writeUtype(loc, val)
			writeItype(loc[4:], val)
		case elf.R_RISCV_TLS_GOT_HI20:
			utils.Write(loc, uint32(sym.GetGotTpAddr(ctx)+A-P))
		case elf.R_RISCV_PCREL_HI20:
			utils.Write(loc, uint32(S+A-P))
		case elf.R_RISCV_HI20:
			writeUtype(loc, uint32(S+A))
		case elf.R_RISCV_LO12_I, elf.R_RISCV_LO12_S:
			val := S + A
			if rel.Type == uint32(elf.R_RISCV_LO12_I) {
				writeItype(loc, uint32(val))
			} else {
				writeStype(loc, uint32(val))
			}

			if utils.SignExtend(val, 11) == val {
				setRs1(loc, 0)
			}
		case elf.R_RISCV_TPREL_LO12_I, elf.R_RISCV_TPREL_LO12_S:
			val := S + A - ctx.MtpAddr
			if rel.Type == uint32(elf.R_RISCV_TPREL_LO12_I) {
				writeItype(loc, uint32(val))
			} else {
				writeStype(loc, uint32(val))
			}

			if utils.SignExtend(val, 11) == val {
				setRs1(loc, 4)
			}
		}
	}

	for a := 0; a < len(rels); a++ {
		switch elf.R_RISCV(rels[a].Type) {
		case elf.R_RISCV_PCREL_LO12_I, elf.R_RISCV_PCREL_LO12_S:
			sym := i.MparentFile.MallSymbols[rels[a].Sym]
			if sym.MinputSection != i {
				utils.FatalExit("Wrong sym.InputSection")
			}
			loc := base[rels[a].Offset:]
			val := utils.BinRead[uint32](base[sym.Mvalue:])

			if rels[a].Type == uint32(elf.R_RISCV_PCREL_LO12_I) {
				writeItype(loc, val)
			} else {
				writeStype(loc, val)
			}
		}
	}

	for a := 0; a < len(rels); a++ {
		switch elf.R_RISCV(rels[a].Type) {
		case elf.R_RISCV_PCREL_HI20, elf.R_RISCV_TLS_GOT_HI20:
			loc := base[rels[a].Offset:]
			val := utils.BinRead[uint32](loc)
			utils.Write(loc, utils.BinRead[uint32](i.Mcontents[rels[a].Offset:]))
			writeUtype(loc, val)
		}
	}
}

func itype(val uint32) uint32 {
	return val << 20
}

func stype(val uint32) uint32 {
	return utils.Bits(val, 11, 5)<<25 | utils.Bits(val, 4, 0)<<7
}

func btype(val uint32) uint32 {
	return utils.Bit(val, 12)<<31 | utils.Bits(val, 10, 5)<<25 |
		utils.Bits(val, 4, 1)<<8 | utils.Bit(val, 11)<<7
}

func utype(val uint32) uint32 {
	return (val + 0x800) & 0xffff_f000
}

func jtype(val uint32) uint32 {
	return utils.Bit(val, 20)<<31 | utils.Bits(val, 10, 1)<<21 |
		utils.Bit(val, 11)<<20 | utils.Bits(val, 19, 12)<<12
}

func cbtype(val uint16) uint16 {
	return utils.Bit(val, 8)<<12 | utils.Bit(val, 4)<<11 | utils.Bit(val, 3)<<10 |
		utils.Bit(val, 7)<<6 | utils.Bit(val, 6)<<5 | utils.Bit(val, 2)<<4 |
		utils.Bit(val, 1)<<3 | utils.Bit(val, 5)<<2
}

func cjtype(val uint16) uint16 {
	return utils.Bit(val, 11)<<12 | utils.Bit(val, 4)<<11 | utils.Bit(val, 9)<<10 |
		utils.Bit(val, 8)<<9 | utils.Bit(val, 10)<<8 | utils.Bit(val, 6)<<7 |
		utils.Bit(val, 7)<<6 | utils.Bit(val, 3)<<5 | utils.Bit(val, 2)<<4 |
		utils.Bit(val, 1)<<3 | utils.Bit(val, 5)<<2
}

func writeItype(loc []byte, val uint32) {
	mask := uint32(0b000000_00000_11111_111_11111_1111111)
	utils.Write(loc, (utils.BinRead[uint32](loc)&mask)|itype(val))
}

func writeStype(loc []byte, val uint32) {
	mask := uint32(0b000000_11111_11111_111_00000_1111111)
	utils.Write(loc, (utils.BinRead[uint32](loc)&mask)|stype(val))
}

func writeBtype(loc []byte, val uint32) {
	mask := uint32(0b000000_11111_11111_111_00000_1111111)
	utils.Write(loc, (utils.BinRead[uint32](loc)&mask)|btype(val))
}

func writeUtype(loc []byte, val uint32) {
	mask := uint32(0b000000_00000_00000_000_11111_1111111)
	utils.Write(loc, (utils.BinRead[uint32](loc)&mask)|utype(val))
}

func writeJtype(loc []byte, val uint32) {
	mask := uint32(0b000000_00000_00000_000_11111_1111111)
	utils.Write(loc, (utils.BinRead[uint32](loc)&mask)|jtype(val))
}

func setRs1(loc []byte, rs1 uint32) {
	utils.Write(loc, utils.BinRead[uint32](loc)&0b111111_11111_00000_111_11111_1111111)
	utils.Write(loc, utils.BinRead[uint32](loc)|(rs1<<15))
}
