package linux_dmabuf

//go:generate go run github.com/rajveermalviya/go-wayland/go-wayland-scanner -pkg linux_dmabuf -prefix zwp -suffix v1 -o linux_dmabuf.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.26/unstable/linux-dmabuf/linux-dmabuf-unstable-v1.xml
