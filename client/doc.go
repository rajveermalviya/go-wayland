// Package client is Go port of wayland-client library
// for writing pure Go GUI software for wayland supported
// platforms.
package client

//go:generate go run ../cmd/go-wayland-scanner/scanner.go -pkg client -prefix wl -o client.go -i https://gitlab.freedesktop.org/wayland/wayland/-/raw/f452e41264387dee4fd737cbf1af58b34b53941b/protocol/wayland.xml
