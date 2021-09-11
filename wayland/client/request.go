package client

import (
	"errors"

	"github.com/rajveermalviya/go-wayland/wayland/internal/byteorder"
)

func (ctx *Context) WriteMsg(b []byte, oob []byte) error {
	n, oobn, err := ctx.Conn.WriteMsgUnix(b, oob, nil)
	if err != nil {
		return err
	}
	if n != len(b) || oobn != len(oob) {
		return errors.New("unable to write request")
	}

	return nil
}

func PutUint32(dst []byte, v uint32) {
	byteorder.NativeEndian.PutUint32(dst, v)
}

func PutString(dst []byte, v string, l int) {
	byteorder.NativeEndian.PutUint32(dst[:4], uint32(l))

	v += "\x00"
	copy(dst[4:4+len(v)], []byte(v))
}

func StringPaddedLen(s string) int {
	stringLen := len(s) + 1
	return (stringLen + (4 - (stringLen & 0x3)))
}
