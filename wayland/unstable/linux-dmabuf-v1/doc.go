package linux_dmabuf

//go:generate go-wayland-scanner -pkg linux_dmabuf -prefix zwp -suffix v1 -o linux_dmabuf.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/f01202f4b73aaf0b9c2c58673d9a932e5a24f054/unstable/linux-dmabuf/linux-dmabuf-unstable-v1.xml
