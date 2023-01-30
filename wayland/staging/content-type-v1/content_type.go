// Generated by go-wayland-scanner
// https://github.com/rajveermalviya/go-wayland/cmd/go-wayland-scanner
// XML file : https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.31/staging/content-type/content-type-v1.xml
//
// content_type_v1 Protocol Copyright:
//
// Copyright © 2021 Emmanuel Gil Peyrot
// Copyright © 2022 Xaver Hugl
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

package content_type

import "github.com/rajveermalviya/go-wayland/wayland/client"

// ContentTypeManager : surface content type manager
//
// This interface allows a client to describe the kind of content a surface
// will display, to allow the compositor to optimize its behavior for it.
//
// Warning! The protocol described in this file is currently in the testing
// phase. Backward compatible changes may be added together with the
// corresponding interface version bump. Backward incompatible changes can
// only be done by creating a new major version of the extension.
type ContentTypeManager struct {
	client.BaseProxy
}

// NewContentTypeManager : surface content type manager
//
// This interface allows a client to describe the kind of content a surface
// will display, to allow the compositor to optimize its behavior for it.
//
// Warning! The protocol described in this file is currently in the testing
// phase. Backward compatible changes may be added together with the
// corresponding interface version bump. Backward incompatible changes can
// only be done by creating a new major version of the extension.
func NewContentTypeManager(ctx *client.Context) *ContentTypeManager {
	wpContentTypeManagerV1 := &ContentTypeManager{}
	ctx.Register(wpContentTypeManagerV1)
	return wpContentTypeManagerV1
}

// Destroy : destroy the content type manager object
//
// Destroy the content type manager. This doesn't destroy objects created
// with the manager.
func (i *ContentTypeManager) Destroy() error {
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

// GetSurfaceContentType : create a new toplevel decoration object
//
// Create a new content type object associated with the given surface.
//
// Creating a wp_content_type_v1 from a wl_surface which already has one
// attached is a client error: already_constructed.
func (i *ContentTypeManager) GetSurfaceContentType(surface *client.Surface) (*ContentType, error) {
	id := NewContentType(i.Context())
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
	client.PutUint32(_reqBuf[l:l+4], surface.ID())
	l += 4
	err := i.Context().WriteMsg(_reqBuf[:], nil)
	return id, err
}

type ContentTypeManagerError uint32

// ContentTypeManagerError :
const (
	// ContentTypeManagerErrorAlreadyConstructed : wl_surface already has a content type object
	ContentTypeManagerErrorAlreadyConstructed ContentTypeManagerError = 0
)

func (e ContentTypeManagerError) Name() string {
	switch e {
	case ContentTypeManagerErrorAlreadyConstructed:
		return "already_constructed"
	default:
		return ""
	}
}

func (e ContentTypeManagerError) Value() string {
	switch e {
	case ContentTypeManagerErrorAlreadyConstructed:
		return "0"
	default:
		return ""
	}
}

func (e ContentTypeManagerError) String() string {
	return e.Name() + "=" + e.Value()
}

// ContentType : content type object for a surface
//
// The content type object allows the compositor to optimize for the kind
// of content shown on the surface. A compositor may for example use it to
// set relevant drm properties like "content type".
//
// The client may request to switch to another content type at any time.
// When the associated surface gets destroyed, this object becomes inert and
// the client should destroy it.
type ContentType struct {
	client.BaseProxy
}

// NewContentType : content type object for a surface
//
// The content type object allows the compositor to optimize for the kind
// of content shown on the surface. A compositor may for example use it to
// set relevant drm properties like "content type".
//
// The client may request to switch to another content type at any time.
// When the associated surface gets destroyed, this object becomes inert and
// the client should destroy it.
func NewContentType(ctx *client.Context) *ContentType {
	wpContentTypeV1 := &ContentType{}
	ctx.Register(wpContentTypeV1)
	return wpContentTypeV1
}

// Destroy : destroy the content type object
//
// Switch back to not specifying the content type of this surface. This is
// equivalent to setting the content type to none, including double
// buffering semantics. See set_content_type for details.
func (i *ContentType) Destroy() error {
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

// SetContentType : specify the content type
//
// Set the surface content type. This informs the compositor that the
// client believes it is displaying buffers matching this content type.
//
// This is purely a hint for the compositor, which can be used to adjust
// its behavior or hardware settings to fit the presented content best.
//
// The content type is double-buffered state, see wl_surface.commit for
// details.
//
//	contentType: the content type
func (i *ContentType) SetContentType(contentType uint32) error {
	const opcode = 1
	const _reqBufLen = 8 + 4
	var _reqBuf [_reqBufLen]byte
	l := 0
	client.PutUint32(_reqBuf[l:4], i.ID())
	l += 4
	client.PutUint32(_reqBuf[l:l+4], uint32(_reqBufLen<<16|opcode&0x0000ffff))
	l += 4
	client.PutUint32(_reqBuf[l:l+4], uint32(contentType))
	l += 4
	err := i.Context().WriteMsg(_reqBuf[:], nil)
	return err
}

type ContentTypeType uint32

// ContentTypeType : possible content types
//
// These values describe the available content types for a surface.
const (
	ContentTypeTypeNone  ContentTypeType = 0
	ContentTypeTypePhoto ContentTypeType = 1
	ContentTypeTypeVideo ContentTypeType = 2
	ContentTypeTypeGame  ContentTypeType = 3
)

func (e ContentTypeType) Name() string {
	switch e {
	case ContentTypeTypeNone:
		return "none"
	case ContentTypeTypePhoto:
		return "photo"
	case ContentTypeTypeVideo:
		return "video"
	case ContentTypeTypeGame:
		return "game"
	default:
		return ""
	}
}

func (e ContentTypeType) Value() string {
	switch e {
	case ContentTypeTypeNone:
		return "0"
	case ContentTypeTypePhoto:
		return "1"
	case ContentTypeTypeVideo:
		return "2"
	case ContentTypeTypeGame:
		return "3"
	default:
		return ""
	}
}

func (e ContentTypeType) String() string {
	return e.Name() + "=" + e.Value()
}