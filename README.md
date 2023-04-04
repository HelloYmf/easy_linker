# easy_linker

## 介绍

一款支持64位RISC-V的静态链接器，目前只支持解析ELF文件格式，后续会加入PE的支持。

## 交叉编译工具
- sudo apt upgrate                                </br>
- sudo apt install qemu-user                      </br>
- sudo apt search risc                            </br>
- sudo apt install gcc-12-riscv64-linux-gnu       </br>
- sudo ln -sf /usr/bin/riscv64-linux-gnu-gcc-12 /usr/bin/riscv64-linux-gnu-gcc                             </br>

## ELF可执行文件结构
```
ELF Header
Program Header
[segment         ][segment                  ][segment                           ]
[section][section][section][section][section][section][section][section][section]
Section Header
```

## linker大致流程
- Command line options* Symbol resolution (including archive processing)
- Process input sections
- Section based garbage collection / identical code folding
- Create synthetic (linker generated) sections
- Scan relocations* Create output sections
- Assign input sections to output sections
- Write file header
- Write sections