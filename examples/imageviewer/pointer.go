package main

import (
	"log"

	"github.com/rajveermalviya/go-wayland/wayland/client"
	"github.com/rajveermalviya/go-wayland/wayland/cursor"
	xdg_shell "github.com/rajveermalviya/go-wayland/wayland/stable/xdg-shell"
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
	pointer.AddEnterHandler(app.HandlePointerEnter)
	pointer.AddLeaveHandler(app.HandlePointerLeave)
	pointer.AddMotionHandler(app.HandlePointerMotion)
	pointer.AddButtonHandler(app.HandlePointerButton)
	pointer.AddAxisHandler(app.HandlePointerAxis)
	pointer.AddAxisSourceHandler(app.HandlePointerAxisSource)
	pointer.AddAxisStopHandler(app.HandlePointerAxisStop)
	pointer.AddAxisDiscreteHandler(app.HandlePointerAxisDiscrete)
	pointer.AddFrameHandler(app.HandlePointerFrame)

	logPrintln("pointer interface registered")
}

func (app *appState) releasePointer() {
	if err := app.pointer.Release(); err != nil {
		logPrintln("unable to release pointer interface")
	}
	app.pointer = nil

	logPrintln("pointer interface released")
}

func (app *appState) HandlePointerEnter(e client.PointerEnterEvent) {
	app.pointerEvent.eventMask |= pointerEventEnter
	app.pointerEvent.serial = e.Serial
	app.pointerEvent.surfaceX = uint32(e.SurfaceX)
	app.pointerEvent.surfaceY = uint32(e.SurfaceY)
}

func (app *appState) HandlePointerLeave(e client.PointerLeaveEvent) {
	app.pointerEvent.eventMask |= pointerEventLeave
	app.pointerEvent.serial = e.Serial
}

func (app *appState) HandlePointerMotion(e client.PointerMotionEvent) {
	app.pointerEvent.eventMask |= pointerEventMotion
	app.pointerEvent.time = e.Time
	app.pointerEvent.surfaceX = uint32(e.SurfaceX)
	app.pointerEvent.surfaceY = uint32(e.SurfaceY)
}

func (app *appState) HandlePointerButton(e client.PointerButtonEvent) {
	app.pointerEvent.eventMask |= pointerEventButton
	app.pointerEvent.serial = e.Serial
	app.pointerEvent.time = e.Time
	app.pointerEvent.button = e.Button
	app.pointerEvent.state = e.State
}

func (app *appState) HandlePointerAxis(e client.PointerAxisEvent) {
	app.pointerEvent.eventMask |= pointerEventAxis
	app.pointerEvent.time = e.Time
	app.pointerEvent.axes[e.Axis].valid = true
	app.pointerEvent.axes[e.Axis].value = int32(e.Value)
}

func (app *appState) HandlePointerAxisSource(e client.PointerAxisSourceEvent) {
	app.pointerEvent.eventMask |= pointerEventAxis
	app.pointerEvent.axisSource = e.AxisSource
}

func (app *appState) HandlePointerAxisStop(e client.PointerAxisStopEvent) {
	app.pointerEvent.eventMask |= pointerEventAxisStop
	app.pointerEvent.time = e.Time
	app.pointerEvent.axes[e.Axis].valid = true
}

func (app *appState) HandlePointerAxisDiscrete(e client.PointerAxisDiscreteEvent) {
	app.pointerEvent.eventMask |= pointerEventAxisDiscrete
	app.pointerEvent.axes[e.Axis].valid = true
	app.pointerEvent.axes[e.Axis].discrete = e.Discrete
}

var cursorMap = map[xdg_shell.ToplevelResizeEdge]string{
	xdg_shell.ToplevelResizeEdgeTop:         cursor.TopSide,
	xdg_shell.ToplevelResizeEdgeTopLeft:     cursor.TopLeftCorner,
	xdg_shell.ToplevelResizeEdgeTopRight:    cursor.TopRightCorner,
	xdg_shell.ToplevelResizeEdgeBottom:      cursor.BottomSide,
	xdg_shell.ToplevelResizeEdgeBottomLeft:  cursor.BottomLeftCorner,
	xdg_shell.ToplevelResizeEdgeBottomRight: cursor.BottomRightCorner,
	xdg_shell.ToplevelResizeEdgeLeft:        cursor.LeftSide,
	xdg_shell.ToplevelResizeEdgeRight:       cursor.RightSide,
	xdg_shell.ToplevelResizeEdgeNone:        cursor.LeftPtr,
}

func (app *appState) HandlePointerFrame(_ client.PointerFrameEvent) {
	e := app.pointerEvent

	if (e.eventMask & pointerEventEnter) != 0 {
		logPrintf("entered %v, %v", e.surfaceX, e.surfaceY)

		edge := componentEdge(uint32(app.width), uint32(app.height), e.surfaceX, e.surfaceY, 8)
		logPrintln(edge.Name())
		cursorName, ok := cursorMap[edge]
		if ok {
			app.setCursor(e.serial, cursorName)
		}
	}

	if (e.eventMask & pointerEventLeave) != 0 {
		logPrintln("leave")

		if err := app.pointer.SetCursor(e.serial, nil, 0, 0); err != nil {
			logPrintln("unable to set cursor")
		}
	}
	if (e.eventMask & pointerEventMotion) != 0 {
		logPrintf("motion %v, %v", e.surfaceX, e.surfaceY)

		edge := componentEdge(uint32(app.width), uint32(app.height), e.surfaceX, e.surfaceY, 8)
		logPrintln(edge.Name())
		cursorName, ok := cursorMap[edge]
		if ok {
			app.setCursor(e.serial, cursorName)
		}
	}
	if (e.eventMask & pointerEventButton) != 0 {
		if e.state == uint32(client.PointerButtonStateReleased) {
			logPrintf("button %d released", e.button)
		} else {
			logPrintf("button %d pressed", e.button)

			switch e.button {
			case BtnLeft:
				edge := componentEdge(uint32(app.width), uint32(app.height), e.surfaceX, e.surfaceY, 8)
				if edge != xdg_shell.ToplevelResizeEdgeNone {
					if err := app.xdgTopLevel.Resize(app.seat, e.serial, uint32(edge)); err != nil {
						logPrintln("unable to start resize")
					}
				} else {
					if err := app.xdgTopLevel.Move(app.seat, e.serial); err != nil {
						logPrintln("unable to start move")
					}
				}
			case BtnRight:
				if err := app.xdgTopLevel.ShowWindowMenu(app.seat, e.serial, int32(e.surfaceX), int32(e.surfaceY)); err != nil {
					logPrintln("unable to show window menu")
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
			logPrintf("%s axis ", client.PointerAxis(i).Name())
			if (e.eventMask & pointerEventAxis) != 0 {
				logPrintf("value %v", e.axes[i].value)
			}
			if (e.eventMask & pointerEventAxisDiscrete) != 0 {
				logPrintf("discrete %d ", e.axes[i].discrete)
			}
			if (e.eventMask & pointerEventAxisSource) != 0 {
				logPrintf("via %s", client.PointerAxisSource(e.axisSource).Name())
			}
			if (e.eventMask & pointerEventAxisStop) != 0 {
				logPrintf("(stopped)")
			}
		}
	}

	// keep surface location in state
	app.pointerEvent = pointerEvent{
		surfaceX: e.surfaceX,
		surfaceY: e.surfaceY,
	}
}

func componentEdge(width, height, pointerX, pointerY, margin uint32) xdg_shell.ToplevelResizeEdge {
	top := pointerY < margin
	bottom := pointerY > (height - margin)
	left := pointerX < margin
	right := pointerX > (width - margin)

	if top {
		if left {
			return xdg_shell.ToplevelResizeEdgeTopLeft
		} else if right {
			return xdg_shell.ToplevelResizeEdgeTopRight
		} else {
			return xdg_shell.ToplevelResizeEdgeTop
		}
	} else if bottom {
		if left {
			return xdg_shell.ToplevelResizeEdgeBottomLeft
		} else if right {
			return xdg_shell.ToplevelResizeEdgeBottomRight
		} else {
			return xdg_shell.ToplevelResizeEdgeBottom
		}
	} else if left {
		return xdg_shell.ToplevelResizeEdgeLeft
	} else if right {
		return xdg_shell.ToplevelResizeEdgeRight
	} else {
		return xdg_shell.ToplevelResizeEdgeNone
	}
}

type cursorData struct {
	name    string
	surface *client.Surface
}

func (c *cursorData) Destory() {
	if err := c.surface.Destroy(); err != nil {
		logPrintln("unable to destory current cursor surface:", err)
	}
	logPrintln("destroyed wl_surface for cursor: ", c.name)
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
	logPrintln("created new wl_surface for cursor: ", cursorName)

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
		logPrintln("unable to set cursor")
	}

	app.currentCursor = &cursorData{name: cursorName, surface: surface}
}
