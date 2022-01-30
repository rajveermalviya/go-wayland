package input_timestamps

//go:generate go-wayland-scanner -pkg input_timestamps -prefix zwp -suffix v1 -o input_timestamps.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.25/unstable/input-timestamps/input-timestamps-unstable-v1.xml
