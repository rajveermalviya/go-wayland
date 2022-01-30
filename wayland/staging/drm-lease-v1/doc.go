package drm_lease

//go:generate go-wayland-scanner -pkg drm_lease -prefix wp -suffix v1 -o drm_lease.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.25/staging/drm-lease/drm-lease-v1.xml
