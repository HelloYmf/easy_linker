# easy_linker

## 交叉编译工具
- sudo apt upgrate                                </br>
- sudo apt install qemu-user                      </br>
- sudo apt search risc                            </br>
- sudo apt install gcc-12-riscv64-linux-gnu       </br>
- sudo ln -sf /usr/bin/riscv64-linux-gnu-gcc-12 /usr/bin/riscv64-linux-gnu-gcc                             </br>

## 进度

- 2023.3.24 完成解析ELF的头部和SectionHeader数组
- 2023.3.25 完成解析符号表和gcc自带链接器命令行参数的模拟解析