package linker

import (
	"debug/elf"

	"github.com/HelloYmf/elf_linker/pkg/utils"
)

// Global offset table
// 库函数中使用到了TLS的东西
type ElfGotSection struct {
	ElfChunk
	MgotTpSyms []*InputElfSymbol
}

func NewGotSection() *ElfGotSection {
	g := &ElfGotSection{
		ElfChunk: NewThunk(),
	}
	g.Mname = ".got"
	g.Mhdr.Type = uint32(elf.SHT_PROGBITS)
	g.Mhdr.Flags = uint64(elf.SHF_ALLOC | elf.SHF_WRITE)
	g.Mhdr.AddrAlign = 0
	return g
}

type GotEntry struct {
	Midx int64
	Mval uint64
}

func (gs *ElfGotSection) AddGotTpSymbol(sym *InputElfSymbol) {
	sym.MgotTpIndx = int32(gs.Mhdr.Size / 8)
	gs.Mhdr.Size += 8
	gs.MgotTpSyms = append(gs.MgotTpSyms, sym)
}

func (gs *ElfGotSection) GetEntries(ctx *LinkContext) []GotEntry {
	entries := make([]GotEntry, 0)
	for _, sym := range gs.MgotTpSyms {
		idx := sym.MgotTpIndx
		entries = append(entries,
			GotEntry{Midx: int64(idx), Mval: sym.GetAddr() - ctx.MtpAddr})
	}

	return entries
}

func (gs *ElfGotSection) CopyBuf(ctx *LinkContext) {
	base := ctx.Mbuf[gs.Mhdr.Offset:]
	for _, ent := range gs.GetEntries(ctx) {
		utils.Write(base[ent.Midx*8:], ent.Mval)
	}
}
