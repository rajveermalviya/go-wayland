package main

import (
	"github.com/rajveermalviya/go-wayland/client"
	"github.com/rajveermalviya/go-wayland/cursor"
	"github.com/rajveermalviya/go-wayland/internal/log"
	xdg_shell "github.com/rajveermalviya/go-wayland/stable/xdg-shell"
)

const (
	pointerEventEnter        = 1 << 0
	pointerEventLeave        = 1 << 1
	pointerEventMotion       = 1 << 2
	pointerEventButton       = 1 << 3
	pointerEventAxis         = 1 << 4
	pointerEventAxisSource   = 1 << 5
	pointerEventAxisStop     = 1 << 6
	pointerEventAxisDiscrete = 1 << 7
)

// From linux/input-event-codes.h
const (
	BtnLeft   = 0x110
	BtnRight  = 0x111
	BtnMiddle = 0x112
)

type pointerEvent struct {
	eventMask          int
	surfaceX, surfaceY uint32
	button, state      uint32
	time               uint32
	serial             uint32
	axes               [2]struct {
		valid    bool
		value    int32
		discrete int32
	}
	axisSource uint32
}

func (app *appState) attachPointer() {
	pointer, err := app.seat.GetPointer()
	if err != nil {
		log.Fatal("unable to register pointer interface")
	}
	app.pointer = pointer
	pointer.AddEnterHandler(app)
	pointer.AddLeaveHandler(app)
	pointer.AddMotionHandler(app)
	pointer.AddButtonHandler(app)
	pointer.AddAxisHandler(app)
	pointer.AddAxisSourceHandler(app)
	pointer.AddAxisStopHandler(app)
	pointer.AddAxisDiscreteHandler(app)
	pointer.AddFrameHandler(app)

	log.Print("pointer interface registered")
}

func (app *appState) releasePointer() {
	app.pointer.RemoveEnterHandler(app)
	app.pointer.RemoveLeaveHandler(app)
	app.pointer.RemoveMotionHandler(app)
	app.pointer.RemoveButtonHandler(app)
	app.pointer.RemoveAxisHandler(app)
	app.pointer.RemoveAxisSourceHandler(app)
	app.pointer.RemoveAxisStopHandler(app)
	app.pointer.RemoveAxisDiscreteHandler(app)
	app.pointer.RemoveFrameHandler(app)

	if err := app.pointer.Release(); err != nil {
		log.Println("unable to release pointer interface")
	}
	app.pointer = nil

	log.Print("pointer interface released")
}

func (app *appState) HandleWlPointerEnter(e client.WlPointerEnterEvent) {
	app.pointerEvent.eventMask |= pointerEventEnter
	app.pointerEvent.serial = e.Serial
	app.pointerEvent.surfaceX = uint32(e.SurfaceX)
	app.pointerEvent.surfaceY = uint32(e.SurfaceY)
}

func (app *appState) HandleWlPointerLeave(e client.WlPointerLeaveEvent) {
	app.pointerEvent.eventMask |= pointerEventLeave
	app.pointerEvent.serial = e.Serial
}

func (app *appState) HandleWlPointerMotion(e client.WlPointerMotionEvent) {
	app.pointerEvent.eventMask |= pointerEventMotion
	app.pointerEvent.time = e.Time
	app.pointerEvent.surfaceX = uint32(e.SurfaceX)
	app.pointerEvent.surfaceY = uint32(e.SurfaceY)
}

func (app *appState) HandleWlPointerButton(e client.WlPointerButtonEvent) {
	app.pointerEvent.eventMask |= pointerEventButton
	app.pointerEvent.serial = e.Serial
	app.pointerEvent.time = e.Time
	app.pointerEvent.button = e.Button
	app.pointerEvent.state = e.State
}

func (app *appState) HandleWlPointerAxis(e client.WlPointerAxisEvent) {
	app.pointerEvent.eventMask |= pointerEventAxis
	app.pointerEvent.time = e.Time
	app.pointerEvent.axes[e.Axis].valid = true
	app.pointerEvent.axes[e.Axis].value = int32(e.Value)
}

func (app *appState) HandleWlPointerAxisSource(e client.WlPointerAxisSourceEvent) {
	app.pointerEvent.eventMask |= pointerEventAxis
	app.pointerEvent.axisSource = e.AxisSource
}

func (app *appState) HandleWlPointerAxisStop(e client.WlPointerAxisStopEvent) {
	app.pointerEvent.eventMask |= pointerEventAxisStop
	app.pointerEvent.time = e.Time
	app.pointerEvent.axes[e.Axis].valid = true
}

func (app *appState) HandleWlPointerAxisDiscrete(e client.WlPointerAxisDiscreteEvent) {
	app.pointerEvent.eventMask |= pointerEventAxisDiscrete
	app.pointerEvent.axes[e.Axis].valid = true
	app.pointerEvent.axes[e.Axis].discrete = e.Discrete
}

var axisName = map[int]string{
	client.WlPointerAxisVerticalScroll:   "vertical",
	client.WlPointerAxisHorizontalScroll: "horizontal",
}

var axisSource = map[uint32]string{
	client.WlPointerAxisSourceWheel:      "wheel",
	client.WlPointerAxisSourceFinger:     "finger",
	client.WlPointerAxisSourceContinuous: "continuous",
	client.WlPointerAxisSourceWheelTilt:  "wheel tilt",
}

var cursorMap = map[uint32]string{
	xdg_shell.XdgToplevelResizeEdgeTop:         cursor.TopSide,
	xdg_shell.XdgToplevelResizeEdgeTopLeft:     cursor.TopLeftCorner,
	xdg_shell.XdgToplevelResizeEdgeTopRight:    cursor.TopRightCorner,
	xdg_shell.XdgToplevelResizeEdgeBottom:      cursor.BottomSide,
	xdg_shell.XdgToplevelResizeEdgeBottomLeft:  cursor.BottomLeftCorner,
	xdg_shell.XdgToplevelResizeEdgeBottomRight: cursor.BottomRightCorner,
	xdg_shell.XdgToplevelResizeEdgeLeft:        cursor.LeftSide,
	xdg_shell.XdgToplevelResizeEdgeRight:       cursor.RightSide,
	xdg_shell.XdgToplevelResizeEdgeNone:        cursor.LeftPtr,
}

func (app *appState) HandleWlPointerFrame(_ client.WlPointerFrameEvent) {
	e := app.pointerEvent

	if (e.eventMask & pointerEventEnter) != 0 {
		log.Printf("entered %v, %v", e.surfaceX, e.surfaceY)

		edge := componentEdge(uint32(app.width), uint32(app.height), e.surfaceX, e.surfaceY, 8)
		cursorName, ok := cursorMap[edge]
		if ok {
			app.setCursor(e.serial, cursorName)
		}
	}

	if (e.eventMask & pointerEventLeave) != 0 {
		log.Print("leave")

		if err := app.pointer.SetCursor(e.serial, nil, 0, 0); err != nil {
			log.Print("unable to set cursor")
		}
	}
	if (e.eventMask & pointerEventMotion) != 0 {
		log.Printf("motion %v, %v", e.surfaceX, e.surfaceY)

		edge := componentEdge(uint32(app.width), uint32(app.height), e.surfaceX, e.surfaceY, 8)
		cursorName, ok := cursorMap[edge]
		if ok {
			app.setCursor(e.serial, cursorName)
		}
	}
	if (e.eventMask & pointerEventButton) != 0 {
		if e.state == client.WlPointerButtonStateReleased {
			log.Printf("button %d released", e.button)
		} else {
			log.Printf("button %d pressed", e.button)

			switch e.button {
			case BtnLeft:
				edge := componentEdge(uint32(app.width), uint32(app.height), e.surfaceX, e.surfaceY, 8)
				if edge != xdg_shell.XdgToplevelResizeEdgeNone {
					if err := app.xdgTopLevel.Resize(app.seat, e.serial, edge); err != nil {
						log.Println("unable to start resize")
					}
				} else {
					if err := app.xdgTopLevel.Move(app.seat, e.serial); err != nil {
						log.Println("unable to start move")
					}
				}
			case BtnRight:
				if err := app.xdgTopLevel.ShowWindowMenu(app.seat, e.serial, int32(e.surfaceX), int32(e.surfaceY)); err != nil {
					log.Println("unable to show window menu")
				}
			}
		}
	}

	const axisEvents = pointerEventAxis | pointerEventAxisSource | pointerEventAxisStop | pointerEventAxisDiscrete

	if (e.eventMask & axisEvents) != 0 {
		for i := 0; i < 2; i++ {
			if !e.axes[i].valid {
				continue
			}
			log.Printf("%s axis ", axisName[i])
			if (e.eventMask & pointerEventAxis) != 0 {
				log.Printf("value %v", e.axes[i].value)
			}
			if (e.eventMask & pointerEventAxisDiscrete) != 0 {
				log.Printf("discrete %d ", e.axes[i].discrete)
			}
			if (e.eventMask & pointerEventAxisSource) != 0 {
				log.Printf("via %s", axisSource[e.axisSource])
			}
			if (e.eventMask & pointerEventAxisStop) != 0 {
				log.Printf("(stopped)")
			}
		}
	}

	// keep surface location in state
	app.pointerEvent = pointerEvent{
		surfaceX: e.surfaceX,
		surfaceY: e.surfaceY,
	}
}

func componentEdge(width, height, pointerX, pointerY, margin uint32) uint32 {
	top := pointerY < margin
	bottom := pointerY > (height - margin)
	left := pointerX < margin
	right := pointerX > (width - margin)

	if top {
		if left {
			log.Print("ToplevelResizeEdgeTopLeft")
			return xdg_shell.XdgToplevelResizeEdgeTopLeft
		} else if right {
			log.Print("ToplevelResizeEdgeTopRight")
			return xdg_shell.XdgToplevelResizeEdgeTopRight
		} else {
			log.Print("ToplevelResizeEdgeTop")
			return xdg_shell.XdgToplevelResizeEdgeTop
		}
	} else if bottom {
		if left {
			log.Print("ToplevelResizeEdgeBottomLeft")
			return xdg_shell.XdgToplevelResizeEdgeBottomLeft
		} else if right {
			log.Print("ToplevelResizeEdgeBottomRight")
			return xdg_shell.XdgToplevelResizeEdgeBottomRight
		} else {
			log.Print("ToplevelResizeEdgeBottom")
			return xdg_shell.XdgToplevelResizeEdgeBottom
		}
	} else if left {
		log.Print("ToplevelResizeEdgeLeft")
		return xdg_shell.XdgToplevelResizeEdgeLeft
	} else if right {
		log.Print("ToplevelResizeEdgeRight")
		return xdg_shell.XdgToplevelResizeEdgeRight
	} else {
		log.Print("ToplevelResizeEdgeNone")
		return xdg_shell.XdgToplevelResizeEdgeNone
	}
}

type cursorData struct {
	name    string
	surface *client.WlSurface
}

func (c *cursorData) Destory() {
	if err := c.surface.Destroy(); err != nil {
		log.Println("unable to destory current cursor surface:", err)
	}
}

func (app *appState) setCursor(serial uint32, cursorName string) {
	if app.currentCursor != nil {
		if cursorName == app.currentCursor.name {
			return
		}

		app.currentCursor.Destory()
	}

	// Get wl_cursor
	c := app.cursorTheme.GetCursor(cursorName)
	if c == nil {
		log.Fatalf("unable to get %v cursor", cursorName)
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

	buffer, err := image.GetBuffer()
	if err != nil {
		log.Fatalf("unable to get image buffer: %v", err)
	}

	// Attach the first image to wl_surface
	if err := surface.Attach(buffer, 0, 0); err != nil {
		log.Fatalf("unable to attach cursor image buffer to cursor suface: %v", err)
	}
	// Commit the surface state changes
	if err2 := surface.Commit(); err2 != nil {
		log.Fatalf("unable to commit surface state: %v", err2)
	}

	if err := app.pointer.SetCursor(
		serial, surface,
		int32(image.HotspotX), int32(image.HotspotY),
	); err != nil {
		log.Print("unable to set cursor")
	}

	app.currentCursor = &cursorData{name: cursorName, surface: surface}
}
