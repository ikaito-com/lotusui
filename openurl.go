package lotusui

import "fmt"

// OpenURL starts the platform URL handler for u (open / xdg-open /
// cmd start). It does not wait for the handler to exit. Empty u is an
// error. No widget and no bundled icon — apps keep their own chrome.
func OpenURL(u string) error {
	if u == "" {
		return fmt.Errorf("lotusui.OpenURL: empty url")
	}
	return openURL(u)
}
