package client

import (
	"fmt"

	"github.com/rajveermalviya/go-wayland/wayland/internal/byteorder"
)

func (ctx *Context) WriteMsg(b []byte, oob []byte) error {
	n, oobn, err := ctx.conn.WriteMsgUnix(b, oob, nil)
	if err != nil {
		return err
	}
	if n != len(b) || oobn != len(oob) {
		return fmt.Errorf("ctx.WriteMsg: incorrect number of bytes written (n=%d oobn=%d)", n, oobn)
	}

	return nil
}

func PutUint32(dst []byte, v uint32) {
	byteorder.NativeEndian.PutUint32(dst, v)
}

func PutFixed(dst []byte, f float64) {
	fx := fixedFromfloat64(f)
	byteorder.NativeEndian.PutUint32(dst, uint32(fx))
}

func PutString(dst []byte, v string, l int) {
	_ = dst[:4+len(v)] // early bounds check

	byteorder.NativeEndian.PutUint32(dst[:4], uint32(l))

	v += "\x00"
	copy(dst[4:4+len(v)], []byte(v))
}

func PaddedLen(l int) int {
	if (l & 0x3) != 0 {
		return l + (4 - (l & 0x3))
	}
	return l
}
