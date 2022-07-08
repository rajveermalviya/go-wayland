package xdg_foreign

//go:generate go run github.com/rajveermalviya/go-wayland/go-wayland-scanner -pkg xdg_foreign -prefix zxdg -suffix v1 -o xdg_foreign.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.26/unstable/xdg-foreign/xdg-foreign-unstable-v1.xml
