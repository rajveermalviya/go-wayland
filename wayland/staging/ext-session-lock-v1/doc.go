package ext_session_lock

//go:generate go-wayland-scanner -pkg ext_session_lock -suffix v1 -o ext_session_lock.go -i https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.25/staging/ext-session-lock/ext-session-lock-v1.xml
