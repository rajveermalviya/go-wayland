package pointer_constraints

//go:generate go run github.com/rajveermalviya/go-wayland/cmd/go-wayland-scanner -pkg pointer_constraints -prefix zwp -suffix v1 -o pointer_constraints.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.31/unstable/pointer-constraints/pointer-constraints-unstable-v1.xml
