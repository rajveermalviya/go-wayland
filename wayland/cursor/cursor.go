// Package cursor is Go port of wayland/cursor library
package cursor

import (
	"errors"
	"os"

	"github.com/rajveermalviya/go-wayland/wayland/client"
	"github.com/rajveermalviya/go-wayland/wayland/cursor/xcursor"
	"github.com/rajveermalviya/go-wayland/wayland/internal/tempfile"
	"golang.org/x/sys/unix"
)

// interesting cursor icons.
const (
	BottomLeftCorner  = "bottom_left_corner"
	BottomRightCorner = "bottom_right_corner"
	BottomSide        = "bottom_side"
	Grabbing          = "grabbing"
	LeftPtr           = "left_ptr"
	LeftSide          = "left_side"
	RightSide         = "right_side"
	TopLeftCorner     = "top_left_corner"
	TopRightCorner    = "top_right_corner"
	TopSide           = "top_side"
	Xterm             = "xterm"
	Hand1             = "hand1"
	Watch             = "watch"
)

type shmPool struct {
	Data []byte

	pool *client.ShmPool
	f    *os.File
	size uint32
	used uint32
}

func createShmPool(shm *client.Shm, size int) (*shmPool, error) {
	f, err := tempfile.Create(int64(size))
	if err != nil {
		return nil, err
	}

	data, err := unix.Mmap(int(f.Fd()), 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	pool, err := shm.CreatePool(int(f.Fd()), int32(size))
	if err != nil {
		return nil, err
	}

	return &shmPool{
		pool: pool,
		f:    f,
		Data: data,
		size: uint32(size),
		used: 0,
	}, nil
}

func (pool *shmPool) Resize(size int) error {
	if err := pool.f.Truncate(int64(size)); err != nil {
		return err
	}

	if err := pool.pool.Resize(int32(size)); err != nil {
		return err
	}

	if err := unix.Munmap(pool.Data); err != nil {
		return err
	}
	pool.Data = nil

	data, err := unix.Mmap(int(pool.f.Fd()), 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return err
	}
	pool.Data = data

	pool.size = uint32(size)

	return nil
}

func (pool *shmPool) Allocate(size int) (int, error) {
	if pool.used+uint32(size) > pool.size {
		if err := pool.Resize(2*int(pool.size) + size); err != nil {
			return 0, err
		}
	}

	offset := pool.used
	pool.used += uint32(size)

	return int(offset), nil
}

func (pool *shmPool) Destroy() error {
	if err := unix.Munmap(pool.Data); err != nil {
		return err
	}
	pool.Data = nil

	if err := pool.pool.Destroy(); err != nil {
		return err
	}

	if err := pool.f.Close(); err != nil {
		return err
	}

	return nil
}

// Image is a still image part of a cursor
//
// Use Image.GetBuffer() to get the corresponding WlBuffer
// to attach to your WlSurface.
type Image struct {
	// Actual Width
	Width uint32

	// Actual Height
	Height uint32

	// Hotspot x (must be inside image)
	HotspotX uint32

	// Hotspot y (must be inside image)
	HotspotY uint32

	// Animation Delay to next frame (ms)
	Delay uint32

	theme  *Theme
	buffer *client.Buffer
	offset int // data offset of this image in the shm pool
}

func (image *Image) GetBuffer() (*client.Buffer, error) {
	theme := image.theme

	if image.buffer == nil {
		buffer, err := theme.pool.pool.CreateBuffer(
			int32(image.offset), int32(image.Width), int32(image.Height),
			int32(image.Width)*4, uint32(client.ShmFormatArgb8888),
		)
		if err != nil {
			return nil, err
		}
		image.buffer = buffer
	}

	return image.buffer, nil
}

func (image *Image) Destroy() error {
	if image.buffer != nil {
		if err := image.buffer.Destroy(); err != nil {
			return err
		}
	}
	return nil
}

// Cursor as returned by Theme.GetCursor()
type Cursor struct {
	// slice of still images composing this animation
	Images []Image

	// name of this cursor
	Name string

	// length of the animation in ms
	totalDelay uint32
}

func createCursorFromXcursorImages(name string, xcimages []xcursor.Image, theme *Theme) (*Cursor, error) {
	images := make([]Image, len(xcimages))
	totalDelay := uint32(0)

	for i, image := range xcimages {
		size := image.Width * image.Height * 4
		offset, err := theme.pool.Allocate(int(size))
		if err != nil {
			return nil, err
		}

		// Copy pixels to shm pool
		copy(theme.pool.Data[offset:], image.Pixels)
		totalDelay += image.Delay

		images[i] = Image{
			theme:    theme,
			Width:    image.Width,
			Height:   image.Height,
			HotspotX: image.HotspotX,
			HotspotY: image.HotspotY,
			Delay:    image.Delay,
			offset:   offset,
		}
	}

	return &Cursor{
		Name:       name,
		totalDelay: totalDelay,
		Images:     images,
	}, nil
}

// FrameAndDuration finds the frame for a given elapsed time in a
// cursor animation as well as the time left until next cursor change.
//
//	cursor: The cursor
//	time: Elapsed time in ms since the beginning of the animation
//	duration: Time left for this image or zero if the cursor won't change.
//
// Returns the index of the image that should be displayed for the
// given time in the cursor animation and updated duration.
func (cursor *Cursor) FrameAndDuration(time uint32, d uint32) (int, uint32) {
	if len(cursor.Images) == 1 || cursor.totalDelay == 0 {
		return 0, 0
	}

	i := 0
	t := time % cursor.totalDelay
	duration := d

	// If there is a 0 delay in the image set then this
	// loop breaks on it and we display that cursor until
	// time % cursor.totalDelay wraps again.
	//
	// Since a 0 delay is silly, and we've never actually
	// seen one in a cursor file, we haven't bothered to
	// "fix" this.
	for t-cursor.Images[i].Delay < t {
		i++
		t -= cursor.Images[i].Delay
	}

	if duration != 0 {
		return i, duration
	}

	// Make sure we don't accidentally tell the caller this is
	// a static cursor image.
	if t >= cursor.Images[i].Delay {
		duration = 1
	} else {
		duration = cursor.Images[i].Delay - t
	}

	return i, duration
}

// Frame finds the frame for a given elapsed time in a cursor animation
//
//	cursor: The cursor
//	time: Elapsed time in ms since the beginning of the animation
//
// Returns the index of the image that should be displayed for the
// given time in the cursor animation.
func (cursor *Cursor) Frame(time uint32) int {
	i, _ := cursor.FrameAndDuration(time, 0)
	return i
}

func (cursor *Cursor) Destroy() error {
	err := MultiError{}

	for _, image := range cursor.Images {
		err.Add(image.Destroy())
	}

	cursor.Images = nil

	return err.Err()
}

type Theme struct {
	cursors map[string]*Cursor
	pool    *shmPool
	size    int
}

func (theme *Theme) loadCallback(name string, images []xcursor.Image) {
	if c := theme.GetCursor(name); c != nil {
		return
	}

	cursor, err := createCursorFromXcursorImages(name, images, theme)
	if err != nil {
		return
	}

	theme.cursors[name] = cursor
}

// LoadTheme loads a cursor theme to memory shared with the compositor
//
// name: The name of the cursor theme to load. If empty, the default theme will be loaded.
// size: Desired size of the cursor images.
// shm: The compositor's shm interface.
//
// Returns an object representing the theme that should be destroyed with
// Theme.Destroy().
func LoadTheme(name string, size int, shm *client.Shm) (*Theme, error) {
	if name == "" {
		name = "default"
	}

	pool, err := createShmPool(shm, size*size*4)
	if err != nil {
		return nil, err
	}

	theme := &Theme{
		size:    size,
		pool:    pool,
		cursors: map[string]*Cursor{},
	}

	xcursor.LoadTheme(name, size, theme.loadCallback)

	if len(theme.cursors) == 0 {
		xcursor.LoadTheme("", size, theme.loadCallback)
	}

	if len(theme.cursors) == 0 {
		_ = pool.Destroy()
		return nil, errors.New("unable to find cursors in specified theme")
	}

	return theme, nil
}

func (theme *Theme) Destroy() error {
	err := MultiError{}

	for _, cursor := range theme.cursors {
		err.Add(cursor.Destroy())
	}

	err.Add(theme.pool.Destroy())

	return err.Err()
}

// GetCursor gets a cursor for a given name from a cursor theme
//
// Returns the theme's cursor of the given name or nil if there is no
// such cursor
func (theme *Theme) GetCursor(name string) *Cursor {
	return theme.cursors[name]
}
