module github.com/rajveermalviya/go-wayland/examples/imageviewer

go 1.16

require (
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/rajveermalviya/go-wayland/wayland v0.0.0-20210715135234-5a594332168d
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0
)

replace github.com/rajveermalviya/go-wayland/wayland => ../../wayland/
