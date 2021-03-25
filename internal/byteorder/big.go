// +build armbe arm64be mips mips64 mips64p32 ppc ppc64 sparc sparc64 s390 s390x

package byteorder

import "encoding/binary"

var NativeEndian = binary.BigEndian
