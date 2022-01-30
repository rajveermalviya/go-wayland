package pointer_gestures

//go:generate go-wayland-scanner -pkg pointer_gestures -prefix zwp -suffix v1 -o pointer_gestures.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.25/unstable/pointer-gestures/pointer-gestures-unstable-v1.xml
