package client

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
)

type Context struct {
	Conn      *net.UnixConn
	objects   map[uint32]Proxy
	currentID uint32
}

func (ctx *Context) Register(p Proxy) {
	ctx.currentID++
	p.SetID(ctx.currentID)
	p.SetContext(ctx)
	ctx.objects[ctx.currentID] = p
}

func (ctx *Context) Unregister(p Proxy) {
	delete(ctx.objects, p.ID())
}

func (ctx *Context) Close() error {
	return ctx.Conn.Close()
}

func (ctx *Context) Dispatch() {
	e, err := ctx.readEvent()
	if err != nil {
		if errors.Is(err, io.EOF) {
			// connection closed
			return
		}

		log.Printf("unable to read event: %v", err)
	}

	proxy, ok := ctx.objects[e.SenderID]
	if ok {
		if dispatcher, ok := proxy.(Dispatcher); ok {
			dispatcher.Dispatch(e)
		} else {
			log.Print("not dispatched")
		}
	} else {
		log.Print("unable to find proxy for proxyID=", e.SenderID)
	}
}

func Connect(addr string) (*Display, error) {
	if addr == "" {
		runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
		if runtimeDir == "" {
			return nil, errors.New("env XDG_RUNTIME_DIR not set")
		}
		if addr == "" {
			addr = os.Getenv("WAYLAND_DISPLAY")
		}
		if addr == "" {
			addr = "wayland-0"
		}
		addr = runtimeDir + "/" + addr
	}

	ctx := &Context{
		objects: map[uint32]Proxy{},
	}

	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: addr, Net: "unix"})
	if err != nil {
		return nil, err
	}
	ctx.Conn = conn

	return NewDisplay(ctx), nil
}
