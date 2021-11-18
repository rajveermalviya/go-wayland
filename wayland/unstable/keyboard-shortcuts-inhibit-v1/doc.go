package keyboard_shortcuts_inhibit

//go:generate go-wayland-scanner -pkg keyboard_shortcuts_inhibit -prefix zwp -suffix v1 -o keyboard_shortcuts_inhibit.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.23/unstable/keyboard-shortcuts-inhibit/keyboard-shortcuts-inhibit-unstable-v1.xml
