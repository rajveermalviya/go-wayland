package client

import (
	"bytes"
	"fmt"
	"math"

	"golang.org/x/sys/unix"

	"github.com/rajveermalviya/go-wayland/internal/byteorder"
)

type Event struct {
	data    *bytes.Buffer
	scms    []unix.SocketControlMessage
	proxyID uint32
	Opcode  uint32
}

func (ctx *Context) readEvent() (*Event, error) {
	header := make([]byte, 8)
	oob := make([]byte, 24)

	n, oobn, _, _, err := ctx.conn.ReadMsgUnix(header, oob)
	if err != nil {
		return nil, err
	}
	if n != 8 {
		return nil, fmt.Errorf("unable to read message header")
	}
	e := &Event{}
	if oobn > 0 {
		if oobn > len(oob) {
			return nil, fmt.Errorf("insufficient control msg buffer")
		}
		scms, err2 := unix.ParseSocketControlMessage(oob)
		if err2 != nil {
			return nil, fmt.Errorf("control message parse error: %w", err)
		}
		e.scms = scms
	}

	headerBuf := bytes.NewBuffer(header)

	e.proxyID = byteorder.NativeEndian.Uint32(headerBuf.Next(4))
	e.Opcode = uint32(byteorder.NativeEndian.Uint16(headerBuf.Next(2)))
	size := uint32(byteorder.NativeEndian.Uint16(headerBuf.Next(2)))

	// subtract 8 bytes from header
	msgSize := int(size) - 8

	data := make([]byte, msgSize)
	n, err = ctx.conn.Read(data)
	if err != nil {
		return nil, err
	}
	if n != msgSize {
		return nil, fmt.Errorf("invalid message size recieved from read")
	}

	e.data = bytes.NewBuffer(data)

	return e, nil
}

func (e *Event) FD() uintptr {
	if len(e.scms) == 0 {
		return 0
	}
	fds, err := unix.ParseUnixRights(&e.scms[0])
	if err != nil {
		panic("unable to parse unix rights")
	}
	return uintptr(fds[0])
}

func (e *Event) Uint32() uint32 {
	buf := e.data.Next(4)
	if len(buf) != 4 {
		panic("unable to read unsigned int")
	}
	return byteorder.NativeEndian.Uint32(buf)
}

func (e *Event) Proxy(ctx *Context) Proxy {
	return ctx.lookupProxy(e.Uint32())
}

func (e *Event) String() string {
	l := int(e.Uint32())
	buf := e.data.Next(l)
	if len(buf) != l {
		panic("unable to read string")
	}
	ret := string(bytes.TrimRight(buf, "\x00"))
	// padding to 32 bit boundary
	if (l & 0x3) != 0 {
		e.data.Next(4 - (l & 0x3))
	}
	return ret
}

func (e *Event) Int32() int32 {
	return int32(e.Uint32())
}

func (e *Event) Float32() float32 {
	return float32(fixedToFloat64(e.Int32()))
}

func (e *Event) Array() []int32 {
	l := int(e.Uint32())
	arr := make([]int32, l/4)
	for i := range arr {
		arr[i] = e.Int32()
	}
	return arr
}

func fixedToFloat64(fixed int32) float64 {
	dat := ((int64(1023 + 44)) << 52) + (1 << 51) + int64(fixed)
	return math.Float64frombits(uint64(dat)) - float64(3<<43)
}
