package linker

import "debug/elf"

type OutputEhdr struct {
	Chunk

	Phdr []Phdr
}

func NewOutputEhdr() *OutputEhdr {
	o := &OutputEhdr{Chunk: NewChunk()}
	o.Shdr.Flags = uint64(elf.SHF_ALLOC)

}

func toPhdrFlags(chunk Chunker) uint32 {
	ret := uint32(elf.PF_R)
	write := chunk.GetShdr().Flags&uint64(elf.SHF_WRITE) != 0
	if write {
		ret |= uint32(elf.PF_W)
	}

	if chunk.GetShdr().Flags&uint64(elf.SHF_EXECINSTR) != 0 {
		ret |= uint32(elf.PF_X)
	}

	return ret
}

func createPhdr(ctx *Context) []Phdr {
	vec := make([]Phdr, 0)
	define := func(typ, flags uint64, minAlign int64, chunk Chunker) {
		vec = append(vec, Phdr{})
		phdr := &vec[len(vec)-1]
		phdr.Type = uint32(typ)
		phdr.Flags = uint32(flags)
		phdr.Align = uint64(math.Max(
			float64(minAlign),
			float64(chunk.GetShdr().AddrAlign)
		))

		phdr.Offset = chunk.GetShdr().Offset
		if chunk.GetShdr().Type == uint32(elf.SHT_NOBITS) {
			phdr.FileSize = 0
		} else {
			phdr.FileSize = chunk.GetShdr().Size
		}

		phdr.VAddr = chunk.GetShdr().Addr
		phdr.PAddr = chunk.GetShdr().Addr
		phdr.MemSize = chunk.GetShdr().Size
	}

	push := func(chunk Chunker) {
		phdr := &vec[len(vec)-1]
		phdr.Align = uint64(math.Max(
			float64(phdr.Align),
			float64(chunk.GetShdr().AddrAlign)
		))

		if chunk.GetShdr().Type != uint32(elf.SHT_NOBITS) {
			phdr.FileSize = chunk.GetShdr().Addr + chunk.GetShdr().Size - phdr.VAddr
		}

		phdr.MemSize = chunk.GetShdr().Addr + chunk.GetShdr().Size - phdr.VAddr
	}

	define(uint64(elf.PT_PHDR), uint64(elf_PF_R), 8, ctx.Phdr)

	isTls := func(chunk Chunker) bool {
		return chunk.GetShdr().Flags & uint64(elf.SHF_TLS) != 0
	}

	isBss := func(chunk Chunker) bool {
		return chunk.GetShdr().Type == uint32(elf.SHT_NOBITS) && !isTls(chunk)
	}

	isNote := func(chunk Chunker) bool {
		shdr := chunk.GetShdr()
		return shdr.Type == uint32(elf.SHT_NOTE) && shdr.Flags & uint64(elf.SHF_ALLOC) != 0
	}

	end := len(ctx.Chunks)
	for i := 0; i < end; {
		first := ctx.Chunks[i]
		i++
		if !isNote(first) {
			continue
		}

		flags := toPhdrFlags(first)
		alignment := first.GetShdr().AddrAlign
		define(uint64(elf.PT_NOTE), uint64(flags), int64(aligment), first)
		for i < end && isNote(ctx.Chunks[i]) && toPhdrFlags(ctx.Chunks[i]) == flags {
			push(ctx.Chunks[i])
			i++
		}
	}

	{
		chunks := make([]Chunker, 0)
	}
}
