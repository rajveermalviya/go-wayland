package input_method

//go:generate go-wayland-scanner -pkg input_method -prefix zwp -suffix v1 -o input_method.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.25/unstable/input-method/input-method-unstable-v1.xml
