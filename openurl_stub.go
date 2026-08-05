//go:build !darwin && !linux && !windows

package lotusui

import "fmt"

func openURL(u string) error {
	return fmt.Errorf("lotusui.OpenURL: unsupported GOOS")
}
