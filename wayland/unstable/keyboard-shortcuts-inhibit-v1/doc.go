package keyboard_shortcuts_inhibit

//go:generate go run github.com/rajveermalviya/go-wayland/cmd/go-wayland-scanner -pkg keyboard_shortcuts_inhibit -prefix zwp -suffix v1 -o keyboard_shortcuts_inhibit.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.31/unstable/keyboard-shortcuts-inhibit/keyboard-shortcuts-inhibit-unstable-v1.xml
