// Package xcursor is Go port of libxcursor functions required by
// wayland/cursor
package xcursor

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const (
	iconDir             = "/usr/X11R6/lib/X11/icons"
	xcursorPath         = "~/.icons:/usr/share/icons:/usr/share/pixmaps:~/.cursors:/usr/share/cursors/xorg-x11:" + iconDir
	xdgDataHomeFallback = "~/.local/share"
	cursorDir           = "/icons"
)

var libraryPath = func() string {
	var envVar string
	var path string

	envVar = os.Getenv("XCURSOR_PATH")
	if envVar != "" {
		path = envVar
	} else {
		envVar = os.Getenv("XDG_DATA_HOME")
		if envVar != "" {
			path = envVar + cursorDir + ":" + xcursorPath
		} else {
			path = xdgDataHomeFallback + cursorDir + ":" + xcursorPath
		}
	}
	return path
}()

func buildThemeDir(dir string, theme string) string {
	if dir == "" || theme == "" {
		return ""
	}

	home := ""
	if strings.HasPrefix(dir, "~") {
		home = os.Getenv("HOME")
		if home == "" {
			return ""
		}
		dir = strings.TrimPrefix(dir, "~")
	}

	return filepath.Join(home, dir, theme)
}

func themeInherits(full string) string {
	if full == "" {
		return ""
	}

	f, err := os.Open(full)
	if err != nil {
		return ""
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()

		if !strings.HasPrefix(line, "Inherits") {
			continue
		}
		chars := line[8:]

		chars = strings.TrimSpace(chars)

		if !strings.HasPrefix(chars, "=") {
			continue
		}
		chars = chars[1:]

		chars = strings.TrimSpace(chars)

		result := strings.NewReplacer(
			" ", "", "\t", "", "\n", "", // Xcursor whitespaces
			";", "", ",", "", // Xcursor separators
		).Replace(chars)

		if result != "" {
			return result
		}
	}

	return ""
}

func loadAllCursorsFromDir(path string, size int, loadCallback func(name string, images []Image)) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, ent := range entries {
		name := ent.Name()
		full := filepath.Join(path, name)

		f, err := os.Open(full)
		if err != nil {
			continue
		}

		images := fileLoadImages(f, size)
		if len(images) > 0 {
			loadCallback(name, images)
		}

		f.Close()
	}

	return nil
}

// LoadTheme loads all the cursor images of a given theme and its
// inherited themes. Each cursor is loaded into []Image
// which is passed to the caller's load callback. If a cursor appears
// more than once across all the inherited themes, the load callback
// will be called multiple times, with possibly different []Image
// which have the same name.
//
//  theme: The name of theme that should be loaded
//  size: The desired size of the cursor images
//  loadCallback: A callback function that will be called
//   for each cursor loaded. The first parameter is name of the cursor
//   representing the loaded cursor and the second is []Image representing
//   the cursor's animated frames, for static ones slice will contain single image.
func LoadTheme(theme string, size int, loadCallback func(name string, images []Image)) {
	if theme == "" {
		theme = "default"
	}

	var inherits string

	for _, path := range strings.Split(libraryPath, ":") {
		dir := buildThemeDir(path, theme)
		if dir == "" {
			continue
		}

		full := filepath.Join(dir, "cursors", "")

		loadAllCursorsFromDir(full, size, loadCallback)

		if inherits == "" {
			full := filepath.Join(dir, "index.theme")
			inherits = themeInherits(full)
		}
	}

	if inherits != "" {
		for _, i := range strings.Split(inherits, ":") {
			LoadTheme(i, size, loadCallback)
		}
	}
}
