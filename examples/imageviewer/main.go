package main

import (
	"flag"
	"image"
	"os"

	syscall "golang.org/x/sys/unix"

	"github.com/nfnt/resize"
	"github.com/rajveermalviya/go-wayland/client"
	"github.com/rajveermalviya/go-wayland/cursor"
	"github.com/rajveermalviya/go-wayland/internal/log"
	"github.com/rajveermalviya/go-wayland/internal/swizzle"
	"github.com/rajveermalviya/go-wayland/internal/tempfile"
	xdg_shell "github.com/rajveermalviya/go-wayland/stable/xdg-shell"
)

func init() {
	flag.Parse()
}

// Global app state
type appState struct {
	appID         string
	title         string
	pImage        *image.RGBA
	width, height int32
	frame         *image.RGBA
	exitChan      chan struct{}

	display    *client.WlDisplay
	registry   *client.WlRegistry
	shm        *client.WlShm
	compositor *client.WlCompositor
	wmBase     *xdg_shell.XdgWmBase
	seat       *client.WlSeat

	surface     *client.WlSurface
	xdgSurface  *xdg_shell.XdgSurface
	xdgTopLevel *xdg_shell.XdgToplevel

	keyboard *client.WlKeyboard
	pointer  *client.WlPointer

	pointerEvent  pointerEvent
	cursorTheme   *cursor.Theme
	cursors       map[string]*cursorData
	currentCursor string
}

func main() {
	if flag.NArg() == 0 {
		log.Fatalf("usage: %s imagefile", os.Args[0])
	}

	const (
		clampedWidth  = 1920
		clampedHeight = 1080
	)

	fileName := flag.Arg(0)

	// Read the image file to *image.RGBA
	pImage, err := rgbaImageFromFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	// Create a proxy image for large images, makes resizing a little better
	if pImage.Rect.Dy() > pImage.Rect.Dx() && pImage.Rect.Dy() > clampedHeight {
		pImage = resize.Resize(0, clampedHeight, pImage, resize.Bilinear).(*image.RGBA)
		log.Print("creating proxy image, resizing by height clamped to", clampedHeight)
	} else if pImage.Rect.Dx() > pImage.Rect.Dy() && pImage.Rect.Dx() > clampedWidth {
		pImage = resize.Resize(clampedWidth, 0, pImage, resize.Bilinear).(*image.RGBA)
		log.Print("creating proxy image, resizing by width clamped to", clampedWidth)
	}

	// Resize again, for first frame
	frameImage := resize.Resize(0, 480, pImage, resize.Bilinear).(*image.RGBA)
	frameRect := frameImage.Bounds()

	app := &appState{
		// Set the title to `cat.jpg - imageview`
		title: fileName + " - imageviewer",
		appID: "imageviewer",
		// Keep proxy image in cache, for use in resizing
		pImage:   pImage,
		width:    int32(frameRect.Dx()),
		height:   int32(frameRect.Dy()),
		frame:    frameImage,
		exitChan: make(chan struct{}),
		cursors:  make(map[string]*cursorData),
	}

	// Connect to wayland server
	display, err := client.Connect("")
	if err != nil {
		log.Fatalf("unable to connect to wayland server: %v", err)
	}
	app.display = display

	display.AddErrorHandler(app)

	// Start other stuff in function for simplicity
	run(app)

	log.Println("closing")

	// Release the pointer if registered
	if app.pointer != nil {
		app.releasePointer()
	}

	// Release the keyboard if registered
	if app.keyboard != nil {
		app.releaseKeyboard()
	}

	// Release wl_seat handlers
	if app.seat != nil {
		app.seat.RemoveCapabilitiesHandler(app)
		app.seat.RemoveNameHandler(app)

		if err := app.seat.Release(); err != nil {
			log.Println("unable to destroy wl_seat:", err)
		}
		app.seat = nil
	}

	// Release xdg_wmbase
	if app.wmBase != nil {
		app.wmBase.RemovePingHandler(app)

		if err := app.wmBase.Destroy(); err != nil {
			log.Println("unable to destroy xdg_wm_base:", err)
		}
		app.wmBase = nil
	}

	for _, c := range app.cursors {
		if err := c.wlCursor.Destroy(); err != nil {
			log.Println("unable to destroy cursor", c.wlCursor.Name, ":", err)
		}

		if err := c.surface.Destroy(); err != nil {
			log.Println("unable to destroy wl_surface of cursor", c.wlCursor.Name, ":", err)
		}
	}

	if app.cursorTheme != nil {
		if err := app.cursorTheme.Destroy(); err != nil {
			log.Println("unable to destroy cursor theme:", err)
		}
	}

	// Close the wayland server connection
	app.Context().Close()
}

func run(app *appState) {
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
	xdgSurface, err := app.wmBase.GetXdgSurface(surface)
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

	// Preload required cursors
	app.loadCursors()

	// Start the dispatch loop
	for {
		select {
		case <-app.exitChan:
			return

		case app.Dispatch() <- struct{}{}:
		}
	}
}

func (app *appState) Dispatch() chan<- struct{} {
	return app.Context().Dispatch()
}

func (app *appState) Context() *client.Context {
	return app.display.Context()
}

func (app *appState) HandleWlRegistryGlobal(ev client.WlRegistryGlobalEvent) {
	log.Printf("discovered an interface: %q\n", ev.Interface)

	switch ev.Interface {
	case "wl_shm":
		shm := client.NewWlShm(app.display.Context())
		err := app.registry.Bind(ev.Name, ev.Interface, ev.Version, shm)
		if err != nil {
			log.Fatalf("unable to bind wl_shm interface: %v", err)
		}
		app.shm = shm
	case "wl_compositor":
		compositor := client.NewWlCompositor(app.display.Context())
		err := app.registry.Bind(ev.Name, ev.Interface, ev.Version, compositor)
		if err != nil {
			log.Fatalf("unable to bind wl_compositor interface: %v", err)
		}
		app.compositor = compositor
	case "xdg_wm_base":
		wmBase := xdg_shell.NewXdgWmBase(app.display.Context())
		err := app.registry.Bind(ev.Name, ev.Interface, ev.Version, wmBase)
		if err != nil {
			log.Fatalf("unable to bind xdg_wm_base interface: %v", err)
		}
		app.wmBase = wmBase
		// Add xdg_wmbase ping handler `app.HandleWmBasePing`
		wmBase.AddPingHandler(app)
	case "wl_seat":
		seat := client.NewWlSeat(app.display.Context())
		err := app.registry.Bind(ev.Name, ev.Interface, ev.Version, seat)
		if err != nil {
			log.Fatalf("unable to bind wl_seat interface: %v", err)
		}
		app.seat = seat
		// Add Keyboard & Pointer handlers
		seat.AddCapabilitiesHandler(app)
		seat.AddNameHandler(app)
	}
}

func (app *appState) HandleXdgSurfaceConfigure(ev xdg_shell.XdgSurfaceConfigureEvent) {
	// Send ack to xdg_surface that we have a frame.
	if err := app.xdgSurface.AckConfigure(ev.Serial); err != nil {
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

func (app *appState) HandleXdgToplevelConfigure(ev xdg_shell.XdgToplevelConfigureEvent) {
	width := ev.Width
	height := ev.Height

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

func (app *appState) loadCursors() {
	// Load default cursor theme
	theme, err := cursor.LoadTheme(24, app.shm)
	if err != nil {
		log.Fatalf("unable to load cursor theme: %v", err)
	}
	app.cursorTheme = theme

	// Create
	for _, name := range []string{
		cursor.BottomLeftCorner,
		cursor.BottomRightCorner,
		cursor.BottomSide,
		cursor.LeftPtr,
		cursor.LeftSide,
		cursor.RightSide,
		cursor.TopLeftCorner,
		cursor.TopRightCorner,
		cursor.TopSide,
	} {
		// Get wl_cursor
		c, err := theme.GetCursor(name)
		if err != nil {
			log.Fatalf("unable to get %v cursor: %v", name, err)
		}

		// Create a wl_surface for cursor
		surface, err := app.compositor.CreateSurface()
		if err != nil {
			log.Fatalf("unable to create compositor surface: %v", err)
		}
		log.Print("created new wl_surface for cursor: ", c.Name)

		// For now get the first image (there are multiple images because of animated cursors)
		// will figure out cursor animation afterwards
		image := c.Images[0]

		// Attach the first image to wl_surface
		if err := surface.Attach(image.Buffer, 0, 0); err != nil {
			log.Fatalf("unable to attach cursor image buffer to cursor suface: %v", err)
		}
		// Commit the surface state changes
		if err2 := surface.Commit(); err2 != nil {
			log.Fatalf("unable to commit surface state: %v", err2)
		}

		// Store the surface for later (immediate) use
		app.cursors[name] = &cursorData{
			wlCursor: c,
			surface:  surface,
		}
	}
}

func (app *appState) drawFrame() *client.WlBuffer {
	log.Print("drawing frame")

	stride := app.width * 4
	size := stride * app.height

	file, err := tempfile.Create(int64(size))
	if err != nil {
		log.Fatalf("unable to create a temporary file: %v", err)
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("unable to create mapping: %v", err)
	}

	pool, err := app.shm.CreatePool(file.Fd(), size)
	if err != nil {
		log.Fatalf("unable to create shm pool: %v", err)
	}

	buf, err := pool.CreateBuffer(0, app.width, app.height, stride, client.WlShmFormatArgb8888)
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
	imgData := make([]byte, len(app.frame.Pix))
	copy(imgData, app.frame.Pix)
	swizzle.BGRA(imgData)

	// Copy image to buffer
	copy(data, imgData)

	if err := syscall.Munmap(data); err != nil {
		log.Printf("unable to delete mapping: %v", err)
	}
	buf.AddReleaseHandler(bufferReleaser{buf: buf})

	log.Print("drawing frame complete")
	return buf
}

type bufferReleaser struct {
	buf *client.WlBuffer
}

func (b bufferReleaser) HandleWlBufferRelease(ev client.WlBufferReleaseEvent) {
	if err := b.buf.Destroy(); err != nil {
		log.Printf("unable to destroy buffer: %v", err)
	}
}

func (app *appState) HandleWlSeatCapabilities(ev client.WlSeatCapabilitiesEvent) {
	havePointer := (ev.Capabilities * client.WlSeatCapabilityPointer) != 0

	if havePointer && app.pointer == nil {
		app.attachPointer()
	} else if !havePointer && app.pointer != nil {
		app.releasePointer()
	}

	haveKeyboard := (ev.Capabilities * client.WlSeatCapabilityKeyboard) != 0

	if haveKeyboard && app.keyboard == nil {
		app.attachKeyboard()
	} else if !haveKeyboard && app.keyboard != nil {
		app.releaseKeyboard()
	}
}

func (*appState) HandleWlSeatName(ev client.WlSeatNameEvent) {
	log.Printf("seat name: %v", ev.Name)
}

// HandleDisplayError handles client.Display errors
func (*appState) HandleWlDisplayError(ev client.WlDisplayErrorEvent) {
	// Just log.Fatal for now
	log.Fatalf("display error event: %v", ev)
}

// HandleWmBasePing handles xdg ping by doing a Pong request
func (app *appState) HandleXdgWmBasePing(ev xdg_shell.XdgWmBasePingEvent) {
	log.Printf("xdg_wmbase ping: serial=%v", ev.Serial)
	app.wmBase.Pong(ev.Serial)
	log.Print("xdg_wmbase pong sent")
}

func (app *appState) HandleXdgToplevelClose(ev xdg_shell.XdgToplevelCloseEvent) {
	close(app.exitChan)
}

type doner struct {
	ch chan client.WlCallbackDoneEvent
}

func (d doner) HandleWlCallbackDone(ev client.WlCallbackDoneEvent) {
	d.ch <- ev
}

func (app *appState) displayRoundTrip() {
	// Get display sync callback
	callback, err := app.display.Sync()
	if err != nil {
		log.Fatalf("unable to get sync callback: %v", err)
	}
	cdeChan := make(chan client.WlCallbackDoneEvent)
	cdeHandler := doner{cdeChan}
	callback.AddDoneHandler(cdeHandler)

	// Wait for callback to return
loop:
	for {
		select {
		case app.Dispatch() <- struct{}{}:
		case <-cdeChan:
			callback.RemoveDoneHandler(cdeHandler)
			close(cdeChan)
			break loop
		}
	}
}
