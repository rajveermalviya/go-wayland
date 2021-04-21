// Package client is Go port of wayland-client library
// for writing pure Go GUI software for wayland supported
// platforms.
package client

//go:generate go run ../cmd/go-wayland-scanner/scanner.go -pkg client -prefix wl_ -i https://gitlab.freedesktop.org/wayland/wayland/-/raw/3bda3d1b4729c8ee7c533520a199611cb841bc8f/protocol/wayland.xml -o client.go
