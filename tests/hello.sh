#!/bin/bash

# $0表示当前sh文件全路径
test_name=$(basename "$0" .sh)
o_path=out/tests/$test_name

mkdir -p "$o_path"

cat <<EOF | $CC -o "$o_path"/"$test_name".o -c -xc -
#include <stdio.h>

int main(void) {
    printf("Hello, World\n");
    return 0;
}
EOF

$CC -B. -static "$o_path"/"$test_name".o -o "$o_path"/"$test_name".out
qemu-riscv64 "$o_path"/"$test_name".out