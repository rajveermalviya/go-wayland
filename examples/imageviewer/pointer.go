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

func (app *appState) HandleWlPointerEnter(ev client.WlPointerEnterEvent) {
	app.pointerEvent.eventMask |= pointerEventEnter
	app.pointerEvent.serial = ev.Serial
	app.pointerEvent.surfaceX = uint32(ev.SurfaceX)
	app.pointerEvent.surfaceY = uint32(ev.SurfaceY)
}

func (app *appState) HandleWlPointerLeave(ev client.WlPointerLeaveEvent) {
	app.pointerEvent.eventMask |= pointerEventLeave
	app.pointerEvent.serial = ev.Serial
}

func (app *appState) HandleWlPointerMotion(ev client.WlPointerMotionEvent) {
	app.pointerEvent.eventMask |= pointerEventMotion
	app.pointerEvent.time = ev.Time
	app.pointerEvent.surfaceX = uint32(ev.SurfaceX)
	app.pointerEvent.surfaceY = uint32(ev.SurfaceY)
}

func (app *appState) HandleWlPointerButton(ev client.WlPointerButtonEvent) {
	app.pointerEvent.eventMask |= pointerEventButton
	app.pointerEvent.serial = ev.Serial
	app.pointerEvent.time = ev.Time
	app.pointerEvent.button = ev.Button
	app.pointerEvent.state = ev.State
}

func (app *appState) HandleWlPointerAxis(ev client.WlPointerAxisEvent) {
	app.pointerEvent.eventMask |= pointerEventAxis
	app.pointerEvent.time = ev.Time
	app.pointerEvent.axes[ev.Axis].valid = true
	app.pointerEvent.axes[ev.Axis].value = int32(ev.Value)
}

func (app *appState) HandleWlPointerAxisSource(ev client.WlPointerAxisSourceEvent) {
	app.pointerEvent.eventMask |= pointerEventAxis
	app.pointerEvent.axisSource = ev.AxisSource
}

func (app *appState) HandleWlPointerAxisStop(ev client.WlPointerAxisStopEvent) {
	app.pointerEvent.eventMask |= pointerEventAxisStop
	app.pointerEvent.time = ev.Time
	app.pointerEvent.axes[ev.Axis].valid = true
}

func (app *appState) HandleWlPointerAxisDiscrete(ev client.WlPointerAxisDiscreteEvent) {
	app.pointerEvent.eventMask |= pointerEventAxisDiscrete
	app.pointerEvent.axes[ev.Axis].valid = true
	app.pointerEvent.axes[ev.Axis].discrete = ev.Discrete
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

func (app *appState) HandleWlPointerFrame(ev client.WlPointerFrameEvent) {
	event := app.pointerEvent

	if (event.eventMask & pointerEventEnter) != 0 {
		log.Printf("entered %v, %v", event.surfaceX, event.surfaceY)

		app.setCursor(event.serial, cursor.LeftPtr)
	}

	if (event.eventMask & pointerEventLeave) != 0 {
		log.Print("leave")
	}
	if (event.eventMask & pointerEventMotion) != 0 {
		log.Printf("motion %v, %v", event.surfaceX, event.surfaceY)

		edge := componentEdge(uint32(app.width), uint32(app.height), event.surfaceX, event.surfaceY, 8)
		cursorName, ok := cursorMap[edge]
		if ok && cursorName != app.currentCursor {
			app.setCursor(event.serial, cursorName)
		}
	}
	if (event.eventMask & pointerEventButton) != 0 {
		if event.state == client.WlPointerButtonStateReleased {
			log.Printf("button %d released", event.button)
		} else {
			log.Printf("button %d pressed", event.button)

			switch event.button {
			case BtnLeft:
				edge := componentEdge(uint32(app.width), uint32(app.height), event.surfaceX, event.surfaceY, 8)
				if edge != xdg_shell.XdgToplevelResizeEdgeNone {
					if err := app.xdgTopLevel.Resize(app.seat, event.serial, edge); err != nil {
						log.Println("unable to start resize")
					}
				} else {
					if err := app.xdgTopLevel.Move(app.seat, event.serial); err != nil {
						log.Println("unable to start move")
					}
				}
			case BtnRight:
				if err := app.xdgTopLevel.ShowWindowMenu(app.seat, event.serial, int32(event.surfaceX), int32(event.surfaceY)); err != nil {
					log.Println("unable to show window menu")
				}
			}
		}
	}

	const axisEvents = pointerEventAxis | pointerEventAxisSource | pointerEventAxisStop | pointerEventAxisDiscrete

	if (event.eventMask & axisEvents) != 0 {
		for i := 0; i < 2; i++ {
			if !event.axes[i].valid {
				continue
			}
			log.Printf("%s axis ", axisName[i])
			if (event.eventMask & pointerEventAxis) != 0 {
				log.Printf("value %v", event.axes[i].value)
			}
			if (event.eventMask & pointerEventAxisDiscrete) != 0 {
				log.Printf("discrete %d ", event.axes[i].discrete)
			}
			if (event.eventMask & pointerEventAxisSource) != 0 {
				log.Printf("via %s", axisSource[event.axisSource])
			}
			if (event.eventMask & pointerEventAxisStop) != 0 {
				log.Printf("(stopped)")
			}
		}
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
	wlCursor *cursor.Cursor

	surface *client.WlSurface
}

func (app *appState) setCursor(serial uint32, cursorName string) {
	c, ok := app.cursors[cursorName]
	if !ok {
		log.Print("unable to get %v cursor", cursorName)
		return
	}

	image := c.wlCursor.Images[0]
	if err := app.pointer.SetCursor(
		serial, c.surface,
		int32(image.HotspotX), int32(image.HotspotY),
	); err != nil {
		log.Print("unable to set cursor")
	}

	app.currentCursor = cursorName
}
