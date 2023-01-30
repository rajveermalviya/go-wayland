package client

import (
	"fmt"
	"unsafe"
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
	_ = dst[3]
	*(*uint32)(unsafe.Pointer(&dst[0])) = v
}

func PutFixed(dst []byte, f float64) {
	fx := fixedFromfloat64(f)
	_ = dst[3]
	*(*int32)(unsafe.Pointer(&dst[0])) = fx
}

func PutString(dst []byte, v string, l int) {
	PutUint32(dst[:4], uint32(l))
	copy(dst[4:], []byte(v))
}

func PutArray(dst []byte, a []byte) {
	PutUint32(dst[:4], uint32(len(a)))
	copy(dst[4:], a)
}
