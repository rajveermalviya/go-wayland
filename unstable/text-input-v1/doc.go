package text_input

//go:generate go run ../../cmd/go-wayland-scanner/scanner.go -pkg text_input -prefix zwp -suffix v1 -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/177ff9119da526462e5d35fbfde6c84794913787/unstable/text-input/text-input-unstable-v1.xml -o text_input.go
