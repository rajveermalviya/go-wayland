package pointer_constraints

//go:generate go-wayland-scanner -pkg pointer_constraints -prefix zwp -suffix v1 -o pointer_constraints.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.25/unstable/pointer-constraints/pointer-constraints-unstable-v1.xml
