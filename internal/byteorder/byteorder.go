package byteorder

import (
	"encoding/binary"
	"unsafe"
)

var NativeEndian binary.ByteOrder

func init() {
	var x uint32 = 0x01020304
	if *(*byte)(unsafe.Pointer(&x)) == 0x01 {
		NativeEndian = binary.BigEndian
	} else {
		NativeEndian = binary.LittleEndian
	}
}
