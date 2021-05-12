package linux_dmabuf

//go:generate go run ../../cmd/go-wayland-scanner/scanner.go -pkg linux_dmabuf -prefix zwp -suffix v1 -o linux_dmabuf.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/177ff9119da526462e5d35fbfde6c84794913787/unstable/linux-dmabuf/linux-dmabuf-unstable-v1.xml
