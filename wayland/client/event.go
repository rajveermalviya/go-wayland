package client

import (
	"fmt"

	"github.com/rajveermalviya/go-wayland/wayland/internal/byteorder"
	"golang.org/x/sys/unix"

	_ "unsafe"
)

var oobSpace = unix.CmsgSpace(4)

func (ctx *Context) ReadMsg() (senderID uint32, opcode uint16, fd uintptr, data []byte, err error) {
	header := make([]byte, 8)
	oob := make([]byte, oobSpace)

	nh, oobn, _, _, err := ctx.conn.ReadMsgUnix(header, oob)
	if err != nil {
		return senderID, opcode, fd, data, err
	}
	if nh != 8 {
		return senderID, opcode, fd, data, fmt.Errorf("ctx.ReadMsg: incorrect number of bytes read for header (n=%d)", nh)
	}

	if oobn > 0 {
		if oobn > len(oob) {
			return senderID, opcode, fd, data, fmt.Errorf("ctx.ReadMsg: incorrect number of bytes read for oob (oobn=%d)", oobn)
		}
		scms, err2 := unix.ParseSocketControlMessage(oob)
		if err2 != nil {
			return senderID, opcode, fd, data, fmt.Errorf("ctx.ReadMsg: unable to parse control message: %w", err)
		}
		if len(scms) > 0 {
			fds, err2 := unix.ParseUnixRights(&scms[0])
			if err2 != nil {
				return senderID, opcode, fd, data, fmt.Errorf("ctx.ReadMsg: unable to parse unix rights: %w", err2)
			}
			if len(fds) > 0 {
				fd = uintptr(fds[0])
			}
		}
	}

	senderID = byteorder.NativeEndian.Uint32(header[:4])
	opcode = byteorder.NativeEndian.Uint16(header[4:6])
	size := byteorder.NativeEndian.Uint16(header[6:8])

	msgSize := int(size) - 8
	data = make([]byte, msgSize)

	nm, err := ctx.conn.Read(data)
	if err != nil {
		return senderID, opcode, fd, data, err
	}
	if int(nm) != msgSize {
		return senderID, opcode, fd, data, fmt.Errorf("ctx.ReadMsg: incorrect number of bytes read for msg (n=%d)", nm)
	}

	return senderID, opcode, fd, data, nil
}

func Uint32(src []byte) uint32 {
	return byteorder.NativeEndian.Uint32(src)
}

//go:linkname String runtime.gostring
func String([]byte) string

func Fixed(src []byte) float64 {
	i := int32(byteorder.NativeEndian.Uint32(src))
	return fixedToFloat64(i)
}
