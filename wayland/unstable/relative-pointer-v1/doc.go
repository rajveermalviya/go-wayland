package relative_pointer

//go:generate go run github.com/rajveermalviya/go-wayland/cmd/go-wayland-scanner -pkg relative_pointer -prefix zwp -suffix v1 -o relative_pointer.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.31/unstable/relative-pointer/relative-pointer-unstable-v1.xml
