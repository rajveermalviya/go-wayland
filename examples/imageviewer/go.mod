module github.com/rajveermalviya/go-wayland/examples/imageviewer

go 1.17

require (
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/rajveermalviya/go-wayland/wayland v0.0.0-20210913062641-fbfc5d1e80b7
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
	golang.org/x/sys v0.0.0-20210915083310-ed5796bab164
)

replace github.com/rajveermalviya/go-wayland/wayland => ../../wayland/
