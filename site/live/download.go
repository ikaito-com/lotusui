package live

import (
	"gioui.org/layout"
	"gioui.org/widget"

	lotusui "github.com/ikaito-com/lotusui"
)

// The installable docsapp. Asset names must match what
// .github/workflows/release.yml uploads: the site links
// releases/latest/download/<asset>, so a rename breaks both sides.
const downloadBase = "https://github.com/ikaito-com/lotusui/releases/latest/download/"

type downloadTarget struct {
	osKey string // visitor-OS key: "macos" | "windows" | "linux"
	label string
	note  string
	asset string
}

var downloadTargets = []downloadTarget{
	{osKey: "macos", label: "Download for macOS", note: "macOS 12+ — universal .app (Apple Silicon and Intel)", asset: "lotusui-docsapp-macos-universal.zip"},
	{osKey: "windows", label: "Download for Windows", note: "Windows — x64", asset: "lotusui-docsapp-windows-amd64.zip"},
	{osKey: "linux", label: "Download for Linux", note: "Linux — x64 (glibc 2.39+)", asset: "lotusui-docsapp-linux-amd64.tar.gz"},
}

// DownloadURL returns the latest-release installable for a visitor-OS
// key, or "" for an unknown key.
func DownloadURL(osKey string) string {
	for _, t := range downloadTargets {
		if t.osKey == osKey {
			return downloadBase + t.asset
		}
	}
	return ""
}

// DownloadLabel returns the download-button wording for a visitor-OS
// key, or "" for an unknown key.
func DownloadLabel(osKey string) string {
	for _, t := range downloadTargets {
		if t.osKey == osKey {
			return t.label
		}
	}
	return ""
}

var downloadBtns [3]widget.Clickable

func desktopDownloadDemo(th *lotusui.Theme, gtx C) D {
	rows := make([]layout.Widget, 0, len(downloadTargets))
	for i := range downloadTargets {
		t := downloadTargets[i]
		if downloadBtns[i].Clicked(gtx) {
			_ = lotusui.OpenURL(downloadBase + t.asset)
		}
		// The note rides IN the row (docsapp embeds suppress section
		// titles), so "universal / x64" survives on the docs page.
		rows = append(rows, section(th, t.note, lotusui.HStack(lotusui.Space.MD,
			lotusui.Button(th, &downloadBtns[i], t.label, lotusui.ButtonProps{Variant: lotusui.ButtonOutline}),
			func(gtx C) D {
				l := lotusui.LabelBody(th, t.note)
				l.Color = th.Palette.FgMuted
				return l.Layout(gtx)
			},
		)))
	}
	return card(th, gtx, rows...)
}
