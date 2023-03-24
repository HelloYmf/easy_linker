#!/bin/bash

# $0表示当前sh文件全路径
test_name=$(basename "$0" .sh)
o_path=out/tests/$test_name

mkdir -p "$o_path"

cat <<EOF | riscv64-linux-gnu-gcc -o "$o_path"/"$test_name".o -c -xc -
#include <stdio.h>

int main(void) {
    return 0;
}
EOF

./jlinker "$o_path"/"$test_name".o