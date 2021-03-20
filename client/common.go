package client

type Dispatcher interface {
	Dispatch(*Event)
}

type ProxyID uint32

type Proxy interface {
	Context() *Context
	SetContext(ctx *Context)
	ID() ProxyID
	SetID(id ProxyID)
}

type BaseProxy struct {
	ctx *Context
	id  ProxyID
}

func (p *BaseProxy) ID() ProxyID {
	return p.id
}

func (p *BaseProxy) SetID(id ProxyID) {
	p.id = id
}

func (p *BaseProxy) Context() *Context {
	return p.ctx
}

func (p *BaseProxy) SetContext(ctx *Context) {
	p.ctx = ctx
}
