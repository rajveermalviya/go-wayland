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
	objects      map[ProxyID]Proxy
	dispatchChan chan struct{}
	exitChan     chan struct{}
	mu           sync.RWMutex
	currentID    ProxyID
}

func (ctx *Context) Register(proxy Proxy) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.currentID++
	proxy.SetID(ctx.currentID)
	proxy.SetContext(ctx)
	ctx.objects[ctx.currentID] = proxy
}

func (ctx *Context) lookupProxy(id ProxyID) Proxy {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	proxy, ok := ctx.objects[id]
	if !ok {
		return nil
	}
	return proxy
}

func (ctx *Context) Unregister(proxy Proxy) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	delete(ctx.objects, proxy.ID())
}

func (ctx *Context) Close() {
	close(ctx.exitChan)
	ctx.conn.Close()
}

func (ctx *Context) Dispatch() chan<- struct{} {
	return ctx.dispatchChan
}

func Connect(addr string) (*WlDisplay, error) {
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
	ctx := &Context{
		objects:      make(map[ProxyID]Proxy),
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

			ev, err := ctx.readEvent()
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

			proxy := ctx.lookupProxy(ev.pid)
			if proxy != nil {
				if dispatcher, ok := proxy.(Dispatcher); ok {
					dispatcher.Dispatch(ev)
				} else {
					log.Print("not dispatched")
				}
			} else {
				log.Print("proxy is nil")
			}

		case <-ctx.exitChan:
			break loop
		}
	}
}
