//go:build darwin || linux

package lotusui

import (
	"os/exec"
	"runtime"
)

func openURL(u string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	return cmd.Start()
}
