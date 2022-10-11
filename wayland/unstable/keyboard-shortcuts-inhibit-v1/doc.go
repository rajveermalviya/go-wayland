package keyboard_shortcuts_inhibit

//go:generate go run github.com/rajveermalviya/go-wayland/go-wayland-scanner -pkg keyboard_shortcuts_inhibit -prefix zwp -suffix v1 -o keyboard_shortcuts_inhibit.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.27/unstable/keyboard-shortcuts-inhibit/keyboard-shortcuts-inhibit-unstable-v1.xml
