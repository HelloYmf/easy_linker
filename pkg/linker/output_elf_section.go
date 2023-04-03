package linker

import (
	"debug/elf"
)

type ElfOutputSection struct {
	ElfChunk
	Members []*InputElfSection
	Midx    uint32
}

func NewElfOutputSection(
	name string, typ uint32, flags uint64, idx uint32) *ElfOutputSection {
	o := &ElfOutputSection{ElfChunk: NewThunk()}
	o.Mname = name
	o.Mhdr.Type = typ
	o.Mhdr.Flags = flags
	o.Midx = idx
	return o
}

func (o *ElfOutputSection) CopyBuf(ctx *LinkContext) {

	var nums uint64 = 0
	for _, num := range o.Members {
		nums += num.MshSize
	}

	if o.Mhdr.Type == uint32(elf.SHT_NOBITS) {
		return
	}

	base := ctx.Mbuf[o.Mhdr.Offset:]
	for _, isec := range o.Members {
		isec.WriteToBuf(ctx, base[isec.Moffset:])
	}

}

func GetOutputSection(
	ctx *LinkContext, name string, typ, flags uint64) *ElfOutputSection {
	name = GetOuputName(name, flags)
	flags = flags &^ uint64(elf.SHF_GROUP) &^
		uint64(elf.SHF_COMPRESSED) &^ uint64(elf.SHF_LINK_ORDER)

	find := func() *ElfOutputSection {
		for _, osec := range ctx.MoutputSections {
			if name == osec.Mname && typ == uint64(osec.Mhdr.Type) &&
				flags == osec.Mhdr.Flags {
				return osec
			}
		}
		return nil
	}

	if osec := find(); osec != nil {
		return osec
	}

	osec := NewElfOutputSection(name, uint32(typ), flags, uint32(len(ctx.MoutputSections)))
	ctx.MoutputSections = append(ctx.MoutputSections, osec)
	return osec
}
