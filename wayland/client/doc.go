// Package client is Go port of wayland-client library
// for writing pure Go GUI software for wayland supported
// platforms.
package client

//go:generate go-wayland-scanner -pkg client -prefix wl -o client.go -i https://raw.githubusercontent.com/wayland-project/wayland/3e897faa29d13bef6f9af31d4f2e89a526e60f4c/protocol/wayland.xml
