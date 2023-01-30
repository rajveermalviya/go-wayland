package main

import (
	"log"

	"github.com/rajveermalviya/go-wayland/wayland/client"
	"golang.org/x/sys/unix"
)

func (app *appState) attachKeyboard() {
	keyboard, err := app.seat.GetKeyboard()
	if err != nil {
		log.Fatal("unable to register keyboard interface")
	}
	app.keyboard = keyboard

	keyboard.AddKeyHandler(app.HandleKeyboardKey)
	keyboard.AddKeymapHandler(app.HandleKeyboardKeymap)

	logPrintln("keyboard interface registered")
}

func (app *appState) releaseKeyboard() {
	if err := app.keyboard.Release(); err != nil {
		logPrintln("unable to release keyboard interface")
	}
	app.keyboard = nil

	logPrintln("keyboard interface released")
}

func (app *appState) HandleKeyboardKey(e client.KeyboardKeyEvent) {
	// close on "esc"
	if e.Key == 1 {
		app.exit = true
	}
}

func (app *appState) HandleKeyboardKeymap(e client.KeyboardKeymapEvent) {
	defer unix.Close(e.Fd)

	// flags := unix.MAP_SHARED
	// if app.seatVersion >= 7 {
	// 	flags = unix.MAP_PRIVATE
	// }

	// buf, err := unix.Mmap(
	// 	e.Fd,
	// 	0,
	// 	int(e.Size),
	// 	unix.PROT_READ,
	// 	flags,
	// )
	// if err != nil {
	// 	log.Printf("failed to mmap keymap: %v\n", err)
	// 	return
	// }
	// defer unix.Munmap(buf)

	// fmt.Println(string(buf))
}
