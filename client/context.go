package client

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

type Context struct {
	conn                  *net.UnixConn
	objects               map[uint32]Proxy
	dispatchChan          chan struct{}
	dispatcherStoppedChan chan struct{}
	mu                    sync.RWMutex
	currentID             uint32
}

func (ctx *Context) Register(p Proxy) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.currentID++
	p.SetID(ctx.currentID)
	p.SetContext(ctx)
	ctx.objects[ctx.currentID] = p
}

func (ctx *Context) lookupProxy(id uint32) Proxy {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	p, ok := ctx.objects[id]
	if !ok {
		return nil
	}
	return p
}

func (ctx *Context) Unregister(p Proxy) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	delete(ctx.objects, p.ID())
}

func (ctx *Context) Close() error {
	close(ctx.dispatchChan)
	<-ctx.dispatcherStoppedChan
	return ctx.conn.Close()
}

func (ctx *Context) Dispatch() chan<- struct{} {
	return ctx.dispatchChan
}

func Connect(addr string) (*WlDisplay, error) {
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
		objects:               make(map[uint32]Proxy),
		currentID:             0,
		dispatchChan:          make(chan struct{}),
		dispatcherStoppedChan: make(chan struct{}),
	}

	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: addr, Net: "unix"})
	if err != nil {
		return nil, err
	}
	ctx.conn = conn

	// dispatch events in separate gorutine
	go ctx.run()

	return NewWlDisplay(ctx), nil
}

func (ctx *Context) run() {
	for range ctx.dispatchChan {
		e, err := ctx.readEvent()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// connection closed
				break
			}

			log.Printf("unable to read event: %v", err)
		}

		proxy := ctx.lookupProxy(e.proxyID)
		if proxy != nil {
			if dispatcher, ok := proxy.(Dispatcher); ok {
				dispatcher.Dispatch(e)
			} else {
				log.Print("not dispatched")
			}
		} else {
			log.Printf("unable to find proxy for proxyID=%d", e.proxyID)
		}
	}

	close(ctx.dispatcherStoppedChan)
}
