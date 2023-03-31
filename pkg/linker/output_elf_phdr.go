package linker

import (
	"debug/elf"
	"math"

	"github.com/HelloYmf/elf_linker/pkg/file/elf_file"
	"github.com/HelloYmf/elf_linker/pkg/utils"
)

type ElfOutputPhdr struct {
	ElfChunk
	Mphdrs []elf_file.ElfProgramHdr
}

func NewElfOutputPhdr() *ElfOutputPhdr {
	o := &ElfOutputPhdr{ElfChunk: NewThunk()}
	o.Mhdr.Flags = uint64(elf.SHF_ALLOC)
	o.Mhdr.AddrAlign = 8
	return o
}

func toPhdrFlags(chunk ElfChunker) uint32 {
	ret := uint32(elf.PF_R)
	write := chunk.GetSHdr().Flags&uint64(elf.SHF_WRITE) != 0
	if write {
		ret |= uint32(elf.PF_W)
	}
	if chunk.GetSHdr().Flags&uint64(elf.SHF_EXECINSTR) != 0 {
		ret |= uint32(elf.PF_X)
	}
	return ret
}

func createPhdrs(ctx *LinkContext) []elf_file.ElfProgramHdr {
	vec := make([]elf_file.ElfProgramHdr, 0)
	define := func(typ, flags uint64, minAlign int64, chunk ElfChunker) {
		vec = append(vec, elf_file.ElfProgramHdr{})
		phdr := &vec[len(vec)-1]
		phdr.Type = uint32(typ)
		phdr.Flags = uint32(flags)
		phdr.Align = uint64(math.Max(
			float64(minAlign),
			float64(chunk.GetSHdr().AddrAlign)))
		phdr.Offset = chunk.GetSHdr().Offset
		if chunk.GetSHdr().Type == uint32(elf.SHT_NOBITS) {
			phdr.FileSize = 0
		} else {
			phdr.FileSize = chunk.GetSHdr().Size
		}
		phdr.VAddr = chunk.GetSHdr().Addr
		phdr.PAddr = chunk.GetSHdr().Addr
		phdr.MemSize = chunk.GetSHdr().Size
	}

	push := func(chunk ElfChunker) {
		phdr := &vec[len(vec)-1]
		phdr.Align = uint64(math.Max(
			float64(phdr.Align),
			float64(chunk.GetSHdr().AddrAlign)))
		if chunk.GetSHdr().Type != uint32(elf.SHT_NOBITS) {
			phdr.FileSize = chunk.GetSHdr().Addr + chunk.GetSHdr().Size - phdr.VAddr
		}
		phdr.MemSize = chunk.GetSHdr().Addr + chunk.GetSHdr().Size - phdr.VAddr
	}

	isTls := func(chunk ElfChunker) bool {
		return chunk.GetSHdr().Flags&uint64(elf.SHF_TLS) != 0
	}

	isBss := func(chunk ElfChunker) bool {
		return chunk.GetSHdr().Type == uint32(elf.SHT_NOBITS) && !isTls(chunk)
	}

	isNote := func(chunk ElfChunker) bool {
		shdr := chunk.GetSHdr()
		return shdr.Type == uint32(elf.SHT_NOTE) &&
			shdr.Flags&uint64(elf.SHF_ALLOC) != 0
	}

	define(uint64(elf.PT_PHDR), uint64(elf.PF_R), 8, ctx.MoutPHdr)

	end := len(ctx.Mchunks)
	for i := 0; i < end; {
		first := ctx.Mchunks[i]
		i++
		if !isNote(first) {
			continue
		}
		flags := toPhdrFlags(first)
		alignment := first.GetSHdr().AddrAlign
		define(uint64(elf.PT_NOTE), uint64(flags), int64(alignment), first)
		for i < end && isNote(ctx.Mchunks[i]) &&
			toPhdrFlags(ctx.Mchunks[i]) == flags {
			push(ctx.Mchunks[i])
			i++
		}
	}

	{
		chunks := make([]ElfChunker, 0)
		for _, chunk := range ctx.Mchunks {
			chunks = append(chunks, chunk)
		}

		chunks = utils.RemoveIf(chunks, func(chunk ElfChunker) bool {
			return isTlsBss(chunk)
		})

		end := len(chunks)
		for i := 0; i < end; {
			first := chunks[i]
			i++

			if first.GetSHdr().Flags&uint64(elf.SHF_ALLOC) == 0 {
				break
			}

			flags := toPhdrFlags(first)
			define(uint64(elf.PT_LOAD), uint64(flags), 4096, first)

			if !isBss(first) {
				for i < end && !isBss(chunks[i]) &&
					toPhdrFlags(chunks[i]) == flags {
					push(chunks[i])
					i++
				}
			}

			for i < end && isBss(chunks[i]) &&
				toPhdrFlags(chunks[i]) == flags {
				push(chunks[i])
				i++
			}
		}
	}

	for i := 0; i < len(ctx.Mchunks); i++ {
		if !isTls(ctx.Mchunks[i]) {
			continue
		}

		define(uint64(elf.PT_TLS), uint64(toPhdrFlags(ctx.Mchunks[i])),
			1, ctx.Mchunks[i])

		i++

		for i < len(ctx.Mchunks) && isTls(ctx.Mchunks[i]) {
			push(ctx.Mchunks[i])
			i++
		}

		phdr := &vec[len(vec)-1]
		ctx.MtpAddr = phdr.VAddr
	}

	return vec
}

func (o *ElfOutputPhdr) UpdateSHdr(ctx *LinkContext) {
	o.Mphdrs = createPhdrs(ctx)
	o.Mhdr.Size = uint64(len(o.Mphdrs)) * uint64(elf_file.ElfProgramHdrSize)
}

func (o *ElfOutputPhdr) CopyBuf(ctx *LinkContext) {
	utils.Write(ctx.Mbuf[o.Mhdr.Offset:], o.Mphdrs)
}
