//go:build windows

package lotusui

import "os/exec"

func openURL(u string) error {
	// start needs an empty title argument when the URL can contain &.
	return exec.Command("cmd", "/c", "start", "", u).Start()
}
