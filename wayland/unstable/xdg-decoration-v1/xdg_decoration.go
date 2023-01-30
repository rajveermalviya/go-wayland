// Generated by go-wayland-scanner
// https://github.com/rajveermalviya/go-wayland/cmd/go-wayland-scanner
// XML file : https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.31/unstable/xdg-decoration/xdg-decoration-unstable-v1.xml
//
// xdg_decoration_unstable_v1 Protocol Copyright:
//
// Copyright © 2018 Simon Ser
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice (including the next
// paragraph) shall be included in all copies or substantial portions of the
// Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package xdg_decoration

import (
	"github.com/rajveermalviya/go-wayland/wayland/client"
	xdg_shell "github.com/rajveermalviya/go-wayland/wayland/stable/xdg-shell"
)

// DecorationManager : window decoration manager
//
// This interface allows a compositor to announce support for server-side
// decorations.
//
// A window decoration is a set of window controls as deemed appropriate by
// the party managing them, such as user interface components used to move,
// resize and change a window's state.
//
// A client can use this protocol to request being decorated by a supporting
// compositor.
//
// If compositor and client do not negotiate the use of a server-side
// decoration using this protocol, clients continue to self-decorate as they
// see fit.
//
// Warning! The protocol described in this file is experimental and
// backward incompatible changes may be made. Backward compatible changes
// may be added together with the corresponding interface version bump.
// Backward incompatible changes are done by bumping the version number in
// the protocol and interface names and resetting the interface version.
// Once the protocol is to be declared stable, the 'z' prefix and the
// version number in the protocol and interface names are removed and the
// interface version number is reset.
type DecorationManager struct {
	client.BaseProxy
}

// NewDecorationManager : window decoration manager
//
// This interface allows a compositor to announce support for server-side
// decorations.
//
// A window decoration is a set of window controls as deemed appropriate by
// the party managing them, such as user interface components used to move,
// resize and change a window's state.
//
// A client can use this protocol to request being decorated by a supporting
// compositor.
//
// If compositor and client do not negotiate the use of a server-side
// decoration using this protocol, clients continue to self-decorate as they
// see fit.
//
// Warning! The protocol described in this file is experimental and
// backward incompatible changes may be made. Backward compatible changes
// may be added together with the corresponding interface version bump.
// Backward incompatible changes are done by bumping the version number in
// the protocol and interface names and resetting the interface version.
// Once the protocol is to be declared stable, the 'z' prefix and the
// version number in the protocol and interface names are removed and the
// interface version number is reset.
func NewDecorationManager(ctx *client.Context) *DecorationManager {
	zxdgDecorationManagerV1 := &DecorationManager{}
	ctx.Register(zxdgDecorationManagerV1)
	return zxdgDecorationManagerV1
}

// Destroy : destroy the decoration manager object
//
// Destroy the decoration manager. This doesn't destroy objects created
// with the manager.
func (i *DecorationManager) Destroy() error {
	defer i.Context().Unregister(i)
	const opcode = 0
	const _reqBufLen = 8
	var _reqBuf [_reqBufLen]byte
	l := 0
	client.PutUint32(_reqBuf[l:4], i.ID())
	l += 4
	client.PutUint32(_reqBuf[l:l+4], uint32(_reqBufLen<<16|opcode&0x0000ffff))
	l += 4
	err := i.Context().WriteMsg(_reqBuf[:], nil)
	return err
}

// GetToplevelDecoration : create a new toplevel decoration object
//
// Create a new decoration object associated with the given toplevel.
//
// Creating an xdg_toplevel_decoration from an xdg_toplevel which has a
// buffer attached or committed is a client error, and any attempts by a
// client to attach or manipulate a buffer prior to the first
// xdg_toplevel_decoration.configure event must also be treated as
// errors.
func (i *DecorationManager) GetToplevelDecoration(toplevel *xdg_shell.Toplevel) (*ToplevelDecoration, error) {
	id := NewToplevelDecoration(i.Context())
	const opcode = 1
	const _reqBufLen = 8 + 4 + 4
	var _reqBuf [_reqBufLen]byte
	l := 0
	client.PutUint32(_reqBuf[l:4], i.ID())
	l += 4
	client.PutUint32(_reqBuf[l:l+4], uint32(_reqBufLen<<16|opcode&0x0000ffff))
	l += 4
	client.PutUint32(_reqBuf[l:l+4], id.ID())
	l += 4
	client.PutUint32(_reqBuf[l:l+4], toplevel.ID())
	l += 4
	err := i.Context().WriteMsg(_reqBuf[:], nil)
	return id, err
}

// ToplevelDecoration : decoration object for a toplevel surface
//
// The decoration object allows the compositor to toggle server-side window
// decorations for a toplevel surface. The client can request to switch to
// another mode.
//
// The xdg_toplevel_decoration object must be destroyed before its
// xdg_toplevel.
type ToplevelDecoration struct {
	client.BaseProxy
	configureHandler ToplevelDecorationConfigureHandlerFunc
}

// NewToplevelDecoration : decoration object for a toplevel surface
//
// The decoration object allows the compositor to toggle server-side window
// decorations for a toplevel surface. The client can request to switch to
// another mode.
//
// The xdg_toplevel_decoration object must be destroyed before its
// xdg_toplevel.
func NewToplevelDecoration(ctx *client.Context) *ToplevelDecoration {
	zxdgToplevelDecorationV1 := &ToplevelDecoration{}
	ctx.Register(zxdgToplevelDecorationV1)
	return zxdgToplevelDecorationV1
}

// Destroy : destroy the decoration object
//
// Switch back to a mode without any server-side decorations at the next
// commit.
func (i *ToplevelDecoration) Destroy() error {
	defer i.Context().Unregister(i)
	const opcode = 0
	const _reqBufLen = 8
	var _reqBuf [_reqBufLen]byte
	l := 0
	client.PutUint32(_reqBuf[l:4], i.ID())
	l += 4
	client.PutUint32(_reqBuf[l:l+4], uint32(_reqBufLen<<16|opcode&0x0000ffff))
	l += 4
	err := i.Context().WriteMsg(_reqBuf[:], nil)
	return err
}

// SetMode : set the decoration mode
//
// Set the toplevel surface decoration mode. This informs the compositor
// that the client prefers the provided decoration mode.
//
// After requesting a decoration mode, the compositor will respond by
// emitting an xdg_surface.configure event. The client should then update
// its content, drawing it without decorations if the received mode is
// server-side decorations. The client must also acknowledge the configure
// when committing the new content (see xdg_surface.ack_configure).
//
// The compositor can decide not to use the client's mode and enforce a
// different mode instead.
//
// Clients whose decoration mode depend on the xdg_toplevel state may send
// a set_mode request in response to an xdg_surface.configure event and wait
// for the next xdg_surface.configure event to prevent unwanted state.
// Such clients are responsible for preventing configure loops and must
// make sure not to send multiple successive set_mode requests with the
// same decoration mode.
//
//	mode: the decoration mode
func (i *ToplevelDecoration) SetMode(mode uint32) error {
	const opcode = 1
	const _reqBufLen = 8 + 4
	var _reqBuf [_reqBufLen]byte
	l := 0
	client.PutUint32(_reqBuf[l:4], i.ID())
	l += 4
	client.PutUint32(_reqBuf[l:l+4], uint32(_reqBufLen<<16|opcode&0x0000ffff))
	l += 4
	client.PutUint32(_reqBuf[l:l+4], uint32(mode))
	l += 4
	err := i.Context().WriteMsg(_reqBuf[:], nil)
	return err
}

// UnsetMode : unset the decoration mode
//
// Unset the toplevel surface decoration mode. This informs the compositor
// that the client doesn't prefer a particular decoration mode.
//
// This request has the same semantics as set_mode.
func (i *ToplevelDecoration) UnsetMode() error {
	const opcode = 2
	const _reqBufLen = 8
	var _reqBuf [_reqBufLen]byte
	l := 0
	client.PutUint32(_reqBuf[l:4], i.ID())
	l += 4
	client.PutUint32(_reqBuf[l:l+4], uint32(_reqBufLen<<16|opcode&0x0000ffff))
	l += 4
	err := i.Context().WriteMsg(_reqBuf[:], nil)
	return err
}

type ToplevelDecorationError uint32

// ToplevelDecorationError :
const (
	// ToplevelDecorationErrorUnconfiguredBuffer : xdg_toplevel has a buffer attached before configure
	ToplevelDecorationErrorUnconfiguredBuffer ToplevelDecorationError = 0
	// ToplevelDecorationErrorAlreadyConstructed : xdg_toplevel already has a decoration object
	ToplevelDecorationErrorAlreadyConstructed ToplevelDecorationError = 1
	// ToplevelDecorationErrorOrphaned : xdg_toplevel destroyed before the decoration object
	ToplevelDecorationErrorOrphaned ToplevelDecorationError = 2
)

func (e ToplevelDecorationError) Name() string {
	switch e {
	case ToplevelDecorationErrorUnconfiguredBuffer:
		return "unconfigured_buffer"
	case ToplevelDecorationErrorAlreadyConstructed:
		return "already_constructed"
	case ToplevelDecorationErrorOrphaned:
		return "orphaned"
	default:
		return ""
	}
}

func (e ToplevelDecorationError) Value() string {
	switch e {
	case ToplevelDecorationErrorUnconfiguredBuffer:
		return "0"
	case ToplevelDecorationErrorAlreadyConstructed:
		return "1"
	case ToplevelDecorationErrorOrphaned:
		return "2"
	default:
		return ""
	}
}

func (e ToplevelDecorationError) String() string {
	return e.Name() + "=" + e.Value()
}

type ToplevelDecorationMode uint32

// ToplevelDecorationMode : window decoration modes
//
// These values describe window decoration modes.
const (
	// ToplevelDecorationModeClientSide : no server-side window decoration
	ToplevelDecorationModeClientSide ToplevelDecorationMode = 1
	// ToplevelDecorationModeServerSide : server-side window decoration
	ToplevelDecorationModeServerSide ToplevelDecorationMode = 2
)

func (e ToplevelDecorationMode) Name() string {
	switch e {
	case ToplevelDecorationModeClientSide:
		return "client_side"
	case ToplevelDecorationModeServerSide:
		return "server_side"
	default:
		return ""
	}
}

func (e ToplevelDecorationMode) Value() string {
	switch e {
	case ToplevelDecorationModeClientSide:
		return "1"
	case ToplevelDecorationModeServerSide:
		return "2"
	default:
		return ""
	}
}

func (e ToplevelDecorationMode) String() string {
	return e.Name() + "=" + e.Value()
}

// ToplevelDecorationConfigureEvent : suggest a surface change
//
// The configure event asks the client to change its decoration mode. The
// configured state should not be applied immediately. Clients must send an
// ack_configure in response to this event. See xdg_surface.configure and
// xdg_surface.ack_configure for details.
//
// A configure event can be sent at any time. The specified mode must be
// obeyed by the client.
type ToplevelDecorationConfigureEvent struct {
	Mode uint32
}
type ToplevelDecorationConfigureHandlerFunc func(ToplevelDecorationConfigureEvent)

// SetConfigureHandler : sets handler for ToplevelDecorationConfigureEvent
func (i *ToplevelDecoration) SetConfigureHandler(f ToplevelDecorationConfigureHandlerFunc) {
	i.configureHandler = f
}

func (i *ToplevelDecoration) Dispatch(opcode uint32, fd int, data []byte) {
	switch opcode {
	case 0:
		if i.configureHandler == nil {
			return
		}
		var e ToplevelDecorationConfigureEvent
		l := 0
		e.Mode = client.Uint32(data[l : l+4])
		l += 4

		i.configureHandler(e)
	}
}
