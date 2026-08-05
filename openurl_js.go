//go:build js

package lotusui

import "syscall/js"

func openURL(u string) error {
	js.Global().Call("open", u, "_blank", "noopener,noreferrer")
	return nil
}
