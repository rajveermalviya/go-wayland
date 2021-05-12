package xdg_decoration

//go:generate go run ../../cmd/go-wayland-scanner/scanner.go -pkg xdg_decoration -prefix zxdg -suffix v1 -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/177ff9119da526462e5d35fbfde6c84794913787/unstable/xdg-decoration/xdg-decoration-unstable-v1.xml -o xdg_decoration.go
