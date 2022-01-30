package relative_pointer

//go:generate go-wayland-scanner -pkg relative_pointer -prefix zwp -suffix v1 -o relative_pointer.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.25/unstable/relative-pointer/relative-pointer-unstable-v1.xml
