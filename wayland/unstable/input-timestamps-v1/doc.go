package input_timestamps

//go:generate go run github.com/rajveermalviya/go-wayland/go-wayland-scanner -pkg input_timestamps -prefix zwp -suffix v1 -o input_timestamps.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.26/unstable/input-timestamps/input-timestamps-unstable-v1.xml
