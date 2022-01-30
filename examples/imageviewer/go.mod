module github.com/rajveermalviya/go-wayland/examples/imageviewer

go 1.17

require (
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/rajveermalviya/go-wayland/wayland v0.0.0-20211210141637-f0b47a9926e9
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410
	golang.org/x/sys v0.0.0-20211210111614-af8b64212486
)

replace github.com/rajveermalviya/go-wayland/wayland => ../../wayland/
