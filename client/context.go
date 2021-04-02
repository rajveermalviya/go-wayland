package client

import (
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/rajveermalviya/go-wayland/internal/log"
)

type Context struct {
	conn         *net.UnixConn
	objects      map[uint32]Proxy
	dispatchChan chan struct{}
	exitChan     chan struct{}
	mu           sync.RWMutex
	currentID    uint32
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

func (ctx *Context) Close() {
	close(ctx.exitChan)
	ctx.conn.Close()
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
		objects:      make(map[uint32]Proxy),
		currentID:    0,
		dispatchChan: make(chan struct{}),
		exitChan:     make(chan struct{}),
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
loop:
	for {
		select {
		case _, ok := <-ctx.dispatchChan:
			if !ok {
				break loop
			}

			e, err := ctx.readEvent()
			if err != nil {
				if errors.Is(err, io.EOF) {
					// connection closed
					break loop
				}

				if strings.Contains(err.Error(), "use of closed network connection") {
					break loop
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
				log.Print("proxy is nil")
			}

		case <-ctx.exitChan:
			close(ctx.dispatchChan)
			break loop
		}
	}
}
