package lotusui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
)

// AlertDialog is the interruption that requires a decision: a Dialog
// that absorbs every outside click and Escape (no backdrop dismissal,
// no corner ✕) and offers exactly Cancel and Action. Poll Confirmed
// and Cancelled each frame while open.
type AlertDialog struct {
	d      Dialog
	cancel widget.Clickable
	action widget.Clickable
}

// AlertDialogProps are the dialog's texts and role.
type AlertDialogProps struct {
	Title       string
	Description string
	// Size picks the dialog width preset (Size2XS…Size2XL).
	// Sizes / Width / Widths mirror Dialog when set.
	Size   Size
	Sizes  ResponsiveSize
	Width  unit.Dp
	Widths ResponsiveDp
	// Media renders above the title — an icon medallion, an image.
	Media       layout.Widget
	Cancel      string // empty = "Cancel"
	Action      string // empty = "Continue"
	Destructive bool   // renders the action destructively
}

// Appear restarts the entrance animation — call on the closed→open
// transition, exactly like Dialog.
func (a *AlertDialog) Appear() { a.d.Appear() }

// Confirmed reports a click on the action button this frame.
func (a *AlertDialog) Confirmed(gtx layout.Context) bool { return a.action.Clicked(gtx) }

// Cancelled reports a click on the cancel button this frame.
func (a *AlertDialog) Cancelled(gtx layout.Context) bool { return a.cancel.Clicked(gtx) }

// Layout renders the dialog at WINDOW constraints, like Dialog.
func (a *AlertDialog) Layout(th *Theme, gtx layout.Context, o AlertDialogProps) layout.Dimensions {
	cancel, action := o.Cancel, o.Action
	if cancel == "" {
		cancel = "Cancel"
	}
	if action == "" {
		action = "Continue"
	}
	a.d.HideClose = true
	a.d.Size = o.Size
	a.d.Sizes = o.Sizes
	a.d.Width = o.Width
	a.d.Widths = o.Widths
	return a.d.Layout(th, gtx, nil, func(gtx layout.Context) layout.Dimensions {
		actionOpts := ButtonProps{}
		if o.Destructive {
			actionOpts.Variant = ButtonDestructive
		}
		rows := []layout.Widget{}
		if o.Media != nil {
			rows = append(rows, o.Media)
		}
		rows = append(rows,
			LabelTitle(th, o.Title).Layout,
			func(gtx layout.Context) layout.Dimensions {
				l := LabelBody(th, o.Description)
				l.Color = th.Palette.FgMuted
				return l.Layout(gtx)
			},
			RightAligned(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(Button(th, &a.cancel, cancel, ButtonProps{Variant: ButtonOutline})),
					layout.Rigid(HSpacer(th.Space.SM)),
					layout.Rigid(Button(th, &a.action, action, actionOpts)),
				)
			}),
		)
		return VStack(th.Space.MD, rows...)(gtx)
	})
}
