package linux_explicit_synchronization

//go:generate go-wayland-scanner -pkg linux_explicit_synchronization -prefix zwp -suffix v1 -o linux_explicit_synchronization.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.25/unstable/linux-explicit-synchronization/linux-explicit-synchronization-unstable-v1.xml
