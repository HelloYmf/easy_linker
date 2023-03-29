package linker

import (
	"debug/elf"
	"strings"
)

var prefixes = []string{
	".text.", ".data.rel.ro.", ".data.", ".rodata.", ".bss.rel.ro.", ".bss.",
	".init_array.", ".fini_array.", ".tbss.", ".tdata.", ".gcc_except_table.",
	".ctors.", ".dtors.",
}

func GetOuputName(name string, flag uint64) string {

	if (name == ".rodata" || strings.HasPrefix(name, ".rodata.")) &&
		flag&uint64(elf.SHF_MERGE) != 0 {
		if flag&uint64(elf.SHF_STRINGS) != 0 {
			return ".rodata.str"
		}
		return "rodata.cst"
	}

	for _, prefix := range prefixes {
		res := prefix[:len(prefix)-1]
		if res == name || strings.HasPrefix(name, prefix) {
			return res
		}
	}
	return ""
}
