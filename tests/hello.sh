#!/bin/bash

# $0表示当前sh文件全路径
test_name=$(basename "$0" .sh)
o_path=out/tests/$test_name

mkdir -p "$o_path"

cat <<EOF | $CC -o "$o_path"/"$test_name".o -c -xc -
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

char* getstr() {
    char* str = (char*)malloc(100);
    strcpy(str, "HelloTest");
    strcat(str, "world");
    return str;
}

int main(void) {
    char* s = getstr();
    printf("123 %s\n\n", getstr());
    return 0;
}
EOF

$CC -B. -static "$o_path"/"$test_name".o -o "$o_path"/"$test_name".out
qemu-riscv64 "$o_path"/"$test_name".out