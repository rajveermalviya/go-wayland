package main

import (
	"image"
	"log"
	"os"

	"github.com/nfnt/resize"
	"github.com/rajveermalviya/go-wayland/examples/imageviewer/internal/swizzle"
	"github.com/rajveermalviya/go-wayland/examples/imageviewer/internal/tempfile"
	"github.com/rajveermalviya/go-wayland/wayland/client"
	"github.com/rajveermalviya/go-wayland/wayland/cursor"
	xdg_shell "github.com/rajveermalviya/go-wayland/wayland/stable/xdg-shell"
	"golang.org/x/sys/unix"
)

// Global app state
type appState struct {
	appID         string
	title         string
	pImage        *image.RGBA
	width, height int32
	frame         *image.RGBA
	exit          bool

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

	// Read the image file to *image.RGBA
	pImage, err := rgbaImageFromFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	// Resize again, for first frame
	frameImage := resize.Resize(0, 480, pImage, resize.NearestNeighbor).(*image.RGBA)
	frameRect := frameImage.Bounds()

	app := &appState{
		// Set the title to `cat.jpg - imageviewer`
		title: fileName + " - imageviewer",
		appID: "imageviewer",
		// Keep proxy image in cache, for use in resizing
		pImage: pImage,
		width:  int32(frameRect.Dx()),
		height: int32(frameRect.Dy()),
		frame:  frameImage,
	}

	app.initWindow()

	// Start the dispatch loop
	for !app.exit {
		app.dispatch()
	}

	logPrintln("closing")
	app.cleanup()
}

func (app *appState) initWindow() {
	// Connect to wayland server
	display, err := client.Connect("")
	if err != nil {
		log.Fatalf("unable to connect to wayland server: %v", err)
	}
	app.display = display

	display.AddErrorHandler(app.HandleDisplayError)

	// Get global interfaces registry
	registry, err := app.display.GetRegistry()
	if err != nil {
		log.Fatalf("unable to get global registry object: %v", err)
	}
	app.registry = registry

	// Add global interfaces registrar handler
	registry.AddGlobalHandler(app.HandleRegistryGlobal)
	// Wait for interfaces to register
	app.displayRoundTrip()
	// Wait for handler events
	app.displayRoundTrip()

	logPrintln("all interfaces registered")

	// Create a wl_surface for toplevel window
	surface, err := app.compositor.CreateSurface()
	if err != nil {
		log.Fatalf("unable to create compositor surface: %v", err)
	}
	app.surface = surface
	logPrintln("created new wl_surface")

	// attach wl_surface to xdg_wmbase to get toplevel
	// handle
	xdgSurface, err := app.xdgWmBase.GetXdgSurface(surface)
	if err != nil {
		log.Fatalf("unable to get xdg_surface: %v", err)
	}
	app.xdgSurface = xdgSurface
	logPrintln("got xdg_surface")

	// Add xdg_surface configure handler `app.HandleSurfaceConfigure`
	xdgSurface.AddConfigureHandler(app.HandleSurfaceConfigure)
	logPrintln("added configure handler")

	// Get toplevel
	xdgTopLevel, err := xdgSurface.GetToplevel()
	if err != nil {
		log.Fatalf("unable to get xdg_toplevel: %v", err)
	}
	app.xdgTopLevel = xdgTopLevel
	logPrintln("got xdg_toplevel")

	// Add xdg_toplevel configure handler for window resizing
	xdgTopLevel.AddConfigureHandler(app.HandleToplevelConfigure)
	// Add xdg_toplevel close handler
	xdgTopLevel.AddCloseHandler(app.HandleToplevelClose)

	// Set title
	if err2 := xdgTopLevel.SetTitle(app.title); err2 != nil {
		log.Fatalf("unable to set toplevel title: %v", err2)
	}
	// Set appID
	if err2 := xdgTopLevel.SetAppId(app.appID); err2 != nil {
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

func (app *appState) dispatch() {
	app.display.Context().Dispatch()
}

func (app *appState) context() *client.Context {
	return app.display.Context()
}

func (app *appState) HandleRegistryGlobal(e client.RegistryGlobalEvent) {
	logPrintf("discovered an interface: %q\n", e.Interface)

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

		shm.AddFormatHandler(app.HandleShmFormat)
	case "xdg_wm_base":
		xdgWmBase := xdg_shell.NewWmBase(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, xdgWmBase)
		if err != nil {
			log.Fatalf("unable to bind xdg_wm_base interface: %v", err)
		}
		app.xdgWmBase = xdgWmBase
		// Add xdg_wmbase ping handler
		xdgWmBase.AddPingHandler(app.HandleWmBasePing)
	case "wl_seat":
		seat := client.NewSeat(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, seat)
		if err != nil {
			log.Fatalf("unable to bind wl_seat interface: %v", err)
		}
		app.seat = seat
		// Add Keyboard & Pointer handlers
		seat.AddCapabilitiesHandler(app.HandleSeatCapabilities)
		seat.AddNameHandler(app.HandleSeatName)
	}
}

func (app *appState) HandleShmFormat(e client.ShmFormatEvent) {
	logPrintf("supported pixel format: %v\n", client.ShmFormat(e.Format))
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
	logPrintln("resizing frame")
	app.frame = resize.Resize(uint(width), uint(height), app.pImage, resize.Bilinear).(*image.RGBA)
	logPrintln("done resizing frame")

	// Update app size
	app.width = width
	app.height = height
}

func (app *appState) drawFrame() *client.Buffer {
	logPrintln("drawing frame")

	stride := app.width * 4
	size := stride * app.height

	file, err := tempfile.Create(int64(size))
	if err != nil {
		log.Fatalf("unable to create a temporary file: %v", err)
	}
	defer func() {
		if err2 := file.Close(); err2 != nil {
			logPrintf("unable to close file: %v", err2)
		}
	}()

	data, err := unix.Mmap(int(file.Fd()), 0, int(size), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		log.Fatalf("unable to create mapping: %v", err)
	}
	defer func() {
		if err2 := unix.Munmap(data); err2 != nil {
			logPrintf("unable to delete mapping: %v", err2)
		}
	}()

	pool, err := app.shm.CreatePool(file.Fd(), size)
	if err != nil {
		log.Fatalf("unable to create shm pool: %v", err)
	}
	defer func() {
		if err2 := pool.Destroy(); err2 != nil {
			logPrintf("unable to destroy shm pool: %v", err2)
		}
	}()

	buf, err := pool.CreateBuffer(0, app.width, app.height, stride, uint32(client.ShmFormatArgb8888))
	if err != nil {
		log.Fatalf("unable to create client.Buffer from shm pool: %v", err)
	}

	// Convert RGBA to BGRA
	copy(data, app.frame.Pix)
	swizzle.BGRA(data)

	buf.AddReleaseHandler(func(_ client.BufferReleaseEvent) {
		if err := buf.Destroy(); err != nil {
			logPrintf("unable to destroy buffer: %v", err)
		}
	})

	logPrintln("drawing frame complete")
	return buf
}

func (app *appState) HandleSeatCapabilities(e client.SeatCapabilitiesEvent) {
	havePointer := (e.Capabilities * uint32(client.SeatCapabilityPointer)) != 0

	if havePointer && app.pointer == nil {
		app.attachPointer()
	} else if !havePointer && app.pointer != nil {
		app.releasePointer()
	}

	haveKeyboard := (e.Capabilities * uint32(client.SeatCapabilityKeyboard)) != 0

	if haveKeyboard && app.keyboard == nil {
		app.attachKeyboard()
	} else if !haveKeyboard && app.keyboard != nil {
		app.releaseKeyboard()
	}
}

func (*appState) HandleSeatName(e client.SeatNameEvent) {
	logPrintf("seat name: %v", e.Name)
}

// HandleDisplayError handles client.Display errors
func (*appState) HandleDisplayError(e client.DisplayErrorEvent) {
	// Just log.Fatal for now
	log.Fatalf("display error event: %v", e)
}

// HandleWmBasePing handles xdg ping by doing a Pong request
func (app *appState) HandleWmBasePing(e xdg_shell.WmBasePingEvent) {
	logPrintf("xdg_wmbase ping: serial=%v", e.Serial)
	app.xdgWmBase.Pong(e.Serial)
	logPrintln("xdg_wmbase pong sent")
}

func (app *appState) HandleToplevelClose(_ xdg_shell.ToplevelCloseEvent) {
	app.exit = true
}

func (app *appState) displayRoundTrip() {
	// Get display sync callback
	callback, err := app.display.Sync()
	if err != nil {
		log.Fatalf("unable to get sync callback: %v", err)
	}
	defer func() {
		if err2 := callback.Destroy(); err2 != nil {
			logPrintln("unable to destroy callback:", err2)
		}
	}()

	done := false
	callback.AddDoneHandler(func(_ client.CallbackDoneEvent) {
		done = true
	})

	// Wait for callback to return
	for !done {
		app.dispatch()
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
			logPrintln("unable to destroy cursor theme:", err)
		}
		app.cursorTheme = nil
	}

	if app.xdgTopLevel != nil {
		if err := app.xdgTopLevel.Destroy(); err != nil {
			logPrintln("unable to destroy xdg_toplevel:", err)
		}
		app.xdgTopLevel = nil
	}

	if app.xdgSurface != nil {
		if err := app.xdgSurface.Destroy(); err != nil {
			logPrintln("unable to destroy xdg_surface:", err)
		}
		app.xdgSurface = nil
	}

	if app.surface != nil {
		if err := app.surface.Destroy(); err != nil {
			logPrintln("unable to destroy wl_surface:", err)
		}
		app.surface = nil
	}

	// Release wl_seat handlers
	if app.seat != nil {
		if err := app.seat.Release(); err != nil {
			logPrintln("unable to destroy wl_seat:", err)
		}
		app.seat = nil
	}

	// Release xdg_wmbase
	if app.xdgWmBase != nil {
		if err := app.xdgWmBase.Destroy(); err != nil {
			logPrintln("unable to destroy xdg_wm_base:", err)
		}
		app.xdgWmBase = nil
	}

	if app.shm != nil {
		if err := app.shm.Destroy(); err != nil {
			logPrintln("unable to destroy wl_shm:", err)
		}
		app.shm = nil
	}

	if app.compositor != nil {
		if err := app.compositor.Destroy(); err != nil {
			logPrintln("unable to destroy wl_compositor:", err)
		}
		app.compositor = nil
	}

	if app.registry != nil {
		if err := app.registry.Destroy(); err != nil {
			logPrintln("unable to destroy wl_registry:", err)
		}
		app.registry = nil
	}

	if app.display != nil {
		if err := app.display.Destroy(); err != nil {
			logPrintln("unable to destroy wl_display:", err)
		}
	}

	// Close the wayland server connection
	if err := app.context().Close(); err != nil {
		logPrintln("unable to close wayland context:", err)
	}
}
