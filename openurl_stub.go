//go:build !darwin && !linux && !windows && !js

package lotusui

import "fmt"

func openURL(u string) error {
	return fmt.Errorf("lotusui.OpenURL: unsupported GOOS")
}
