package tablet

//go:generate go-wayland-scanner -pkg tablet -prefix zwp -suffix v2 -o tablet.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.25/unstable/tablet/tablet-unstable-v2.xml
