package main

import (
	"image"
	"os"

	"golang.org/x/sys/unix"

	"github.com/nfnt/resize"
	"github.com/rajveermalviya/go-wayland/client"
	"github.com/rajveermalviya/go-wayland/cursor"
	"github.com/rajveermalviya/go-wayland/internal/log"
	"github.com/rajveermalviya/go-wayland/internal/swizzle"
	"github.com/rajveermalviya/go-wayland/internal/tempfile"
	xdg_shell "github.com/rajveermalviya/go-wayland/stable/xdg-shell"
)

// Global app state
type appState struct {
	appID         string
	title         string
	pImage        *image.RGBA
	width, height int32
	frame         *image.RGBA
	exitChan      chan struct{}

	display    *client.Display
	registry   *client.Registry
	shm        *client.Shm
	compositor *client.Compositor
	xdgWmBase  *xdg_shell.WmBase
	seat       *client.Seat

	surface     *client.Surface
	xdgSurface  *xdg_shell.Surface
	xdgTopLevel *xdg_shell.Toplevel

	keyboard *client.Keyboard
	pointer  *client.Pointer

	pointerEvent  pointerEvent
	cursorTheme   *cursor.Theme
	currentCursor *cursorData
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s file.jpg", os.Args[0])
	}

	fileName := os.Args[1]

	const (
		maxWidth  = 1920
		maxHeight = 1080
	)

	// Read the image file to *image.RGBA
	pImage, err := rgbaImageFromFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	// Create a proxy image for large images, makes resizing a little better
	if pImage.Rect.Dy() > pImage.Rect.Dx() && pImage.Rect.Dy() > maxHeight {
		pImage = resize.Resize(0, maxHeight, pImage, resize.Bilinear).(*image.RGBA)
		log.Print("creating proxy image, resizing by height clamped to", maxHeight)
	} else if pImage.Rect.Dx() > pImage.Rect.Dy() && pImage.Rect.Dx() > maxWidth {
		pImage = resize.Resize(maxWidth, 0, pImage, resize.Bilinear).(*image.RGBA)
		log.Print("creating proxy image, resizing by width clamped to", maxWidth)
	}

	// Resize again, for first frame
	frameImage := resize.Resize(0, 480, pImage, resize.Bilinear).(*image.RGBA)
	frameRect := frameImage.Bounds()

	app := &appState{
		// Set the title to `cat.jpg - imageviewer`
		title: fileName + " - imageviewer",
		appID: "imageviewer",
		// Keep proxy image in cache, for use in resizing
		pImage:   pImage,
		width:    int32(frameRect.Dx()),
		height:   int32(frameRect.Dy()),
		frame:    frameImage,
		exitChan: make(chan struct{}),
	}

	app.initWindow()

	// Start the dispatch loop
loop:
	for {
		select {
		case <-app.exitChan:
			break loop
		case app.dispatch() <- struct{}{}:
		}
	}

	log.Println("closing")
	app.cleanup()
}

func (app *appState) initWindow() {
	// Connect to wayland server
	display, err := client.Connect("")
	if err != nil {
		log.Fatalf("unable to connect to wayland server: %v", err)
	}
	app.display = display

	display.AddErrorHandler(app)

	// Get global interfaces registry
	registry, err := app.display.GetRegistry()
	if err != nil {
		log.Fatalf("unable to get global registry object: %v", err)
	}
	app.registry = registry

	// Add global interfaces registrar handler
	registry.AddGlobalHandler(app)
	// Wait for interfaces to register
	app.displayRoundTrip()
	// Wait for handler events
	app.displayRoundTrip()

	log.Print("all interfaces registered")

	// Create a wl_surface for toplevel window
	surface, err := app.compositor.CreateSurface()
	if err != nil {
		log.Fatalf("unable to create compositor surface: %v", err)
	}
	app.surface = surface
	log.Print("created new wl_surface")

	// attach wl_surface to xdg_wmbase to get toplevel
	// handle
	xdgSurface, err := app.xdgWmBase.GetXdgSurface(surface)
	if err != nil {
		log.Fatalf("unable to get xdg_surface: %v", err)
	}
	app.xdgSurface = xdgSurface
	log.Print("got xdg_surface")

	// Add xdg_surface configure handler `app.HandleSurfaceConfigure`
	xdgSurface.AddConfigureHandler(app)
	log.Print("added configure handler")

	// Get toplevel
	xdgTopLevel, err := xdgSurface.GetToplevel()
	if err != nil {
		log.Fatalf("unable to get xdg_toplevel: %v", err)
	}
	app.xdgTopLevel = xdgTopLevel
	log.Print("got xdg_toplevel")

	// Add xdg_toplevel configure handler for window resizing
	xdgTopLevel.AddConfigureHandler(app)
	// Add xdg_toplevel close handler
	xdgTopLevel.AddCloseHandler(app)

	// Set title
	if err2 := xdgTopLevel.SetTitle(app.title); err2 != nil {
		log.Fatalf("unable to set toplevel title: %v", err2)
	}
	// Set appID
	if err2 := xdgTopLevel.SetAppID(app.appID); err2 != nil {
		log.Fatalf("unable to set toplevel appID: %v", err2)
	}
	// Commit the state changes (title & appID) to the server
	if err2 := app.surface.Commit(); err2 != nil {
		log.Fatalf("unable to commit surface state: %v", err2)
	}

	// Load default cursor theme
	theme, err := cursor.LoadTheme("default", 24, app.shm)
	if err != nil {
		log.Fatalf("unable to load cursor theme: %v", err)
	}
	app.cursorTheme = theme
}

func (app *appState) dispatch() chan<- struct{} {
	return app.context().Dispatch()
}

func (app *appState) context() *client.Context {
	return app.display.Context()
}

func (app *appState) HandleRegistryGlobal(e client.RegistryGlobalEvent) {
	log.Printf("discovered an interface: %q\n", e.Interface)

	switch e.Interface {
	case "wl_compositor":
		compositor := client.NewCompositor(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, compositor)
		if err != nil {
			log.Fatalf("unable to bind wl_compositor interface: %v", err)
		}
		app.compositor = compositor
	case "wl_shm":
		shm := client.NewShm(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, shm)
		if err != nil {
			log.Fatalf("unable to bind wl_shm interface: %v", err)
		}
		app.shm = shm

		shm.AddFormatHandler(app)
	case "xdg_wm_base":
		xdgWmBase := xdg_shell.NewWmBase(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, xdgWmBase)
		if err != nil {
			log.Fatalf("unable to bind xdg_wm_base interface: %v", err)
		}
		app.xdgWmBase = xdgWmBase
		// Add xdg_wmbase ping handler `app.HandleWmBasePing`
		xdgWmBase.AddPingHandler(app)
	case "wl_seat":
		seat := client.NewSeat(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, seat)
		if err != nil {
			log.Fatalf("unable to bind wl_seat interface: %v", err)
		}
		app.seat = seat
		// Add Keyboard & Pointer handlers
		seat.AddCapabilitiesHandler(app)
		seat.AddNameHandler(app)
	}
}

func (app *appState) HandleShmFormat(e client.ShmFormatEvent) {
	log.Printf("supported pixel format: 0x%08x\n", e.Format)
}

func (app *appState) HandleSurfaceConfigure(e xdg_shell.SurfaceConfigureEvent) {
	// Send ack to xdg_surface that we have a frame.
	if err := app.xdgSurface.AckConfigure(e.Serial); err != nil {
		log.Fatal("unable to ack xdg surface configure")
	}

	// Draw frame
	buffer := app.drawFrame()

	// Attach new frame to the surface
	if err := app.surface.Attach(buffer, 0, 0); err != nil {
		log.Fatalf("unable to attach buffer to surface: %v", err)
	}
	// Commit the surface state
	if err := app.surface.Commit(); err != nil {
		log.Fatalf("unable to commit surface state: %v", err)
	}
}

func (app *appState) HandleToplevelConfigure(e xdg_shell.ToplevelConfigureEvent) {
	width := e.Width
	height := e.Height

	if width == 0 || height == 0 {
		// Compositor is deferring to us
		return
	}

	if width == app.width && height == app.height {
		// No need to resize
		return
	}

	// Resize the proxy image to new frame size
	// and set it to frame image
	log.Print("resizing frame")
	app.frame = resize.Resize(uint(width), uint(height), app.pImage, resize.Bilinear).(*image.RGBA)
	log.Print("done resizing frame")

	// Update app size
	app.width = width
	app.height = height
}

func (app *appState) drawFrame() *client.Buffer {
	log.Print("drawing frame")

	stride := app.width * 4
	size := stride * app.height

	file, err := tempfile.Create(int64(size))
	if err != nil {
		log.Fatalf("unable to create a temporary file: %v", err)
	}

	data, err := unix.Mmap(int(file.Fd()), 0, int(size), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		log.Fatalf("unable to create mapping: %v", err)
	}

	pool, err := app.shm.CreatePool(file.Fd(), size)
	if err != nil {
		log.Fatalf("unable to create shm pool: %v", err)
	}

	buf, err := pool.CreateBuffer(0, app.width, app.height, stride, client.ShmFormatArgb8888)
	if err != nil {
		log.Fatalf("unable to create client.Buffer from shm pool: %v", err)
	}
	if err := pool.Destroy(); err != nil {
		log.Printf("unable to destroy shm pool: %v", err)
	}
	if err := file.Close(); err != nil {
		log.Printf("unable to close file: %v", err)
	}

	// Convert RGBA to BGRA
	imgData := make([]byte, size)
	copy(imgData, app.frame.Pix)
	swizzle.BGRA(imgData)

	// Copy image to buffer
	copy(data, imgData)

	if err := unix.Munmap(data); err != nil {
		log.Printf("unable to delete mapping: %v", err)
	}
	buf.AddReleaseHandler(bufferReleaser{buf: buf})

	log.Print("drawing frame complete")
	return buf
}

type bufferReleaser struct {
	buf *client.Buffer
}

func (b bufferReleaser) HandleBufferRelease(e client.BufferReleaseEvent) {
	if err := b.buf.Destroy(); err != nil {
		log.Printf("unable to destroy buffer: %v", err)
	}
}

func (app *appState) HandleSeatCapabilities(e client.SeatCapabilitiesEvent) {
	havePointer := (e.Capabilities * client.SeatCapabilityPointer) != 0

	if havePointer && app.pointer == nil {
		app.attachPointer()
	} else if !havePointer && app.pointer != nil {
		app.releasePointer()
	}

	haveKeyboard := (e.Capabilities * client.SeatCapabilityKeyboard) != 0

	if haveKeyboard && app.keyboard == nil {
		app.attachKeyboard()
	} else if !haveKeyboard && app.keyboard != nil {
		app.releaseKeyboard()
	}
}

func (*appState) HandleSeatName(e client.SeatNameEvent) {
	log.Printf("seat name: %v", e.Name)
}

// HandleDisplayError handles client.Display errors
func (*appState) HandleDisplayError(e client.DisplayErrorEvent) {
	// Just log.Fatal for now
	log.Fatalf("display error event: %v", e)
}

// HandleWmBasePing handles xdg ping by doing a Pong request
func (app *appState) HandleWmBasePing(e xdg_shell.WmBasePingEvent) {
	log.Printf("xdg_wmbase ping: serial=%v", e.Serial)
	app.xdgWmBase.Pong(e.Serial)
	log.Print("xdg_wmbase pong sent")
}

func (app *appState) HandleToplevelClose(_ xdg_shell.ToplevelCloseEvent) {
	close(app.exitChan)
}

type doner struct {
	ch chan client.CallbackDoneEvent
}

func (d doner) HandleCallbackDone(e client.CallbackDoneEvent) {
	close(d.ch)
}

func (app *appState) displayRoundTrip() {
	// Get display sync callback
	callback, err := app.display.Sync()
	if err != nil {
		log.Fatalf("unable to get sync callback: %v", err)
	}
	doneChan := make(chan client.CallbackDoneEvent)
	cdeHandler := doner{doneChan}
	callback.AddDoneHandler(cdeHandler)

	// Wait for callback to return
loop:
	for {
		select {
		case app.dispatch() <- struct{}{}:
		case <-doneChan:
			_ = callback.Destroy()
			break loop
		}
	}
}

func (app *appState) cleanup() {
	// Release the pointer if registered
	if app.pointer != nil {
		app.releasePointer()
	}

	// Release the keyboard if registered
	if app.keyboard != nil {
		app.releaseKeyboard()
	}

	if app.currentCursor != nil {
		app.currentCursor.Destory()
		app.currentCursor = nil
	}

	if app.cursorTheme != nil {
		if err := app.cursorTheme.Destroy(); err != nil {
			log.Println("unable to destroy cursor theme:", err)
		}
		app.cursorTheme = nil
	}

	if app.xdgTopLevel != nil {
		if err := app.xdgTopLevel.Destroy(); err != nil {
			log.Println("unable to destroy xdg_toplevel:", err)
		}
		app.xdgTopLevel = nil
	}

	if app.xdgSurface != nil {
		if err := app.xdgSurface.Destroy(); err != nil {
			log.Println("unable to destroy xdg_surface:", err)
		}
		app.xdgSurface = nil
	}

	if app.surface != nil {
		if err := app.surface.Destroy(); err != nil {
			log.Println("unable to destroy wl_surface:", err)
		}
		app.surface = nil
	}

	// Release wl_seat handlers
	if app.seat != nil {
		if err := app.seat.Release(); err != nil {
			log.Println("unable to destroy wl_seat:", err)
		}
		app.seat = nil
	}

	// Release xdg_wmbase
	if app.xdgWmBase != nil {
		if err := app.xdgWmBase.Destroy(); err != nil {
			log.Println("unable to destroy xdg_wm_base:", err)
		}
		app.xdgWmBase = nil
	}

	if app.shm != nil {
		if err := app.shm.Destroy(); err != nil {
			log.Println("unable to destroy wl_shm:", err)
		}
		app.shm = nil
	}

	if app.compositor != nil {
		if err := app.compositor.Destroy(); err != nil {
			log.Println("unable to destroy wl_compositor:", err)
		}
		app.compositor = nil
	}

	if app.registry != nil {
		if err := app.registry.Destroy(); err != nil {
			log.Println("unable to destroy wl_registry:", err)
		}
		app.registry = nil
	}

	if app.display != nil {
		if err := app.display.Destroy(); err != nil {
			log.Println("unable to destroy wl_display:", err)
		}
	}

	// Close the wayland server connection
	if err := app.context().Close(); err != nil {
		log.Println("unable to close wayland context:", err)
	}
}
