module github.com/rajveermalviya/go-wayland/examples

go 1.16

replace github.com/rajveermalviya/go-wayland/wayland => ../wayland/

require (
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/rajveermalviya/go-wayland/wayland v0.0.0-00010101000000-000000000000
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
)
