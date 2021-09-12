//go:build 386 || amd64 || amd64p32 || alpha || arm || arm64 || mipsle || mips64le || mips64p32le || nios2 || ppc64le || riscv || riscv64 || sh || wasm
// +build 386 amd64 amd64p32 alpha arm arm64 mipsle mips64le mips64p32le nios2 ppc64le riscv riscv64 sh wasm

// from golang.org/x/sys

package byteorder

import "encoding/binary"

var NativeEndian = binary.LittleEndian
