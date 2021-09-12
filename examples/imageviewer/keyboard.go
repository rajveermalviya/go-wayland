package main

import (
	"log"

	"github.com/rajveermalviya/go-wayland/wayland/client"
)

func (app *appState) attachKeyboard() {
	keyboard, err := app.seat.GetKeyboard()
	if err != nil {
		log.Fatal("unable to register keyboard interface")
	}
	app.keyboard = keyboard

	keyboard.AddKeyHandler(app)

	logPrintln("keyboard interface registered")
}

func (app *appState) releaseKeyboard() {
	app.keyboard.RemoveKeyHandler(app)

	if err := app.keyboard.Release(); err != nil {
		logPrintln("unable to release keyboard interface")
	}
	app.keyboard = nil

	logPrintln("keyboard interface released")
}

func (app *appState) HandleKeyboardKey(e client.KeyboardKeyEvent) {
	// close on "q"
	if e.Key == 16 {
		app.exit = true
	}
}
