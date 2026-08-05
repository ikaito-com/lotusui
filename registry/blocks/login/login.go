// Package login is the login-form block: a ready-made composition of
// Card, Field, Input and Button — vendor it with `lotusui add
// login-form` and shape it to your app. Blocks use only lotusui's
// exported API, so the vendored copy is a plain file in your package.
package login

import (
	"gioui.org/layout"
	"gioui.org/widget"

	lotusui "github.com/ikaito-com/lotusui"
)

// Form holds the block's interactive state. One Form per on-screen
// instance — identity in immediate mode is the struct.
type Form struct {
	Email    lotusui.Input
	Password lotusui.Input
	Submit   widget.Clickable

	// Busy renders the submit button in its loading state; the caller
	// flips it around the auth round-trip.
	Busy bool
}

// Submitted reports a click on the submit button this frame.
func (f *Form) Submitted(gtx layout.Context) bool { return f.Submit.Clicked(gtx) }

// Layout renders the form as an elevated card: email and password
// fields with labels, and a full-width primary action.
func (f *Form) Layout(th *lotusui.Theme) layout.Widget {
	f.Password.Editor.Mask = '•'
	return lotusui.Card(th, lotusui.CardProps{Variant: lotusui.CardElevated}, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(lotusui.Field(th, lotusui.FieldProps{Label: "Email"}, func(gtx layout.Context) layout.Dimensions {
				return f.Email.LayoutField(th, gtx, "you@example.com")
			})),
			layout.Rigid(lotusui.Spacer(th.Space.MD)),
			layout.Rigid(lotusui.Field(th, lotusui.FieldProps{Label: "Password"}, func(gtx layout.Context) layout.Dimensions {
				return f.Password.LayoutField(th, gtx, "")
			})),
			layout.Rigid(lotusui.Spacer(th.Space.LG)),
			layout.Rigid(lotusui.FullWidth(lotusui.Button(th, &f.Submit, "Sign in", lotusui.ButtonProps{
				Loading: f.Busy, LoadingText: "Signing in…",
			}))),
		)
	})
}
