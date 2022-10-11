package linux_explicit_synchronization

//go:generate go run github.com/rajveermalviya/go-wayland/go-wayland-scanner -pkg linux_explicit_synchronization -prefix zwp -suffix v1 -o linux_explicit_synchronization.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.27/unstable/linux-explicit-synchronization/linux-explicit-synchronization-unstable-v1.xml
