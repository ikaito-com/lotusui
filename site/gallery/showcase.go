package main

// Showcase scenes: composed arrangements of REAL components, designed
// to be screenshotted for the README and the docs homepage (make
// media). They are ordinary gallery states — when components change,
// re-capturing keeps every marketing image truthful.

import (
	"image"
	"image/color"

	"gioui.org/op"
	"gioui.org/unit"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"

	lotusui "github.com/ikaito-com/lotusui"
)

var sc struct {
	email, pw     lotusui.Input
	login, google widget.Clickable
	signup        widget.Clickable
	notif, sec    lotusui.Switch
	fmtGroup      lotusui.ToggleGroup
	vol           lotusui.Slider
	terms         lotusui.Checkbox
	inited        bool
}

func showcaseInit() {
	if sc.inited {
		return
	}
	sc.inited = true
	sc.email.Editor.SetText("m@example.com")
	sc.notif.Value = true
	sc.fmtGroup.Options = markOpts()
	sc.fmtGroup.SetValue("bold")
	sc.vol.Value = 0.62
	sc.terms.Value = true
}

// showcaseDemo is the hero scene: the login card beside a column of
// living controls — one glance says "a complete design system".
func showcaseDemo(th *lotusui.Theme, gtx C) D {
	showcaseInit()
	if sc.terms.Clicked(gtx) {
		sc.terms.Value = !sc.terms.Value
	}
	pad := func(w layout.Widget) layout.Widget {
		return func(gtx C) D {
			// The design language's canvas: white cards floating on the
			// tinted page — the screenshot shows the theme, not a void.
			lotusui.Fill(gtx, th.Palette.Bg)
			return layout.UniformInset(lotusui.Space.LG).Layout(gtx, w)
		}
	}
	loginCard := func(gtx C) D {
		return lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
			func(gtx C) D {
				return layout.Flex{}.Layout(gtx,
					layout.Flexed(1, lotusui.VStack(4,
						lotusui.LabelCardTitle(th, "Login to your account").Layout,
						func(gtx C) D {
							l := lotusui.LabelCaption(th, "Enter your email below to login")
							l.Color = th.Palette.FgMuted
							return l.Layout(gtx)
						},
					)),
					layout.Rigid(lotusui.Button(th, &sc.signup, "Sign Up", lotusui.ButtonProps{Variant: lotusui.ButtonLink, Size: lotusui.SizeSM})),
				)
			},
			lotusui.Field(th, lotusui.FieldProps{Label: "Email"}, func(gtx C) D {
				return sc.email.LayoutField(th, gtx, "m@example.com")
			}),
			lotusui.Field(th, lotusui.FieldProps{Label: "Password"}, func(gtx C) D {
				sc.pw.Editor.Mask = '•'
				return sc.pw.LayoutField(th, gtx, "")
			}),
			func(gtx C) D { return sc.terms.Layout(th, gtx, "Remember me") },
			lotusui.FullWidth(lotusui.Button(th, &sc.login, "Login", lotusui.ButtonProps{})),
			lotusui.FullWidth(lotusui.Button(th, &sc.google, "Login with Google", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})),
		))(gtx)
	}
	row := func(s *lotusui.Switch, title, sub string) layout.Widget {
		return func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, lotusui.VStack(2,
					lotusui.LabelBody(th, title).Layout,
					func(gtx C) D {
						l := lotusui.LabelCaption(th, sub)
						l.Color = th.Palette.FgSubtle
						return l.Layout(gtx)
					},
				)),
				layout.Rigid(func(gtx C) D { return s.Layout(th, gtx) }),
			)
		}
	}
	switchCard := func(gtx C) D {
		return lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
			row(&sc.notif, "Notifications", "Email me about activity."),
			lotusui.Hairline(th),
			row(&sc.sec, "Security emails", "Sign-ins from new devices."),
		))(gtx)
	}
	controlCard := func(gtx C) D {
		return lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(th.Space.MD,
			lotusui.HStack(th.Space.SM,
				lotusui.Badge(th, "v0.1.0", lotusui.BadgeProps{}),
				lotusui.Badge(th, "Verified", lotusui.BadgeProps{Variant: lotusui.BadgeSecondary, Icon: lotusui.IconAccept}),
				lotusui.Badge(th, "Teal", lotusui.BadgeProps{Color: lotusui.Teal}),
			),
			func(gtx C) D {
				return sc.fmtGroup.Layout(th, gtx, lotusui.SizeMD)
			},
			func(gtx C) D { return sc.vol.Layout(th, gtx) },
			lotusui.Progress(th, 0.66),
			lotusui.HStack(-8,
				lotusui.Avatar(th, lotusui.AvatarProps{Initials: "CN"}),
				lotusui.Avatar(th, lotusui.AvatarProps{Initials: "ER", Color: lotusui.Teal}),
				lotusui.Avatar(th, lotusui.AvatarProps{Initials: "LR", Color: lotusui.Pink}),
				lotusui.Avatar(th, lotusui.AvatarProps{}),
			),
		))(gtx)
	}
	return pad(func(gtx C) D {
		return layout.Center.Layout(gtx, func(gtx C) D {
			// The hero IS a Grid demo: the login card spans both rows
			// (RowSpan: 2), so the grid's row contract lands its bottom
			// flush with the second card's — rows equalize to their
			// tallest, spanning items stretch, exactly as documented.
			gtx.Constraints.Max.X = gtx.Dp(780)
			return lotusui.Grid{Columns: 2, Gap: th.Space.LG}.Layout(th, gtx,
				lotusui.GridItem{RowSpan: 2, W: loginCard},
				lotusui.Cell(switchCard),
				lotusui.Cell(controlCard),
			)
		})
	})(gtx)
}

// showcaseColorsDemo is the color-engine strip: one anchor per row,
// the whole interaction ladder derived — buttons, soft badges,
// avatars and the scale steps, per stock scale.
func showcaseColorsDemo(th *lotusui.Theme, gtx C) D {
	scales := []struct {
		name  string
		scale lotusui.ColorScale
	}{
		{"Teal", lotusui.Teal},
		{"Purple", lotusui.Purple},
		{"Pink", lotusui.Pink},
		{"Orange", lotusui.Orange},
		{"Green", lotusui.Green},
	}
	var rows []layout.Widget
	for i, s := range scales {
		s := s
		btn := &showcaseColorBtns[i]
		rows = append(rows, func(gtx C) D {
			// Fixed-width columns: the badge and avatar cells reserve
			// their column regardless of label width, so the swatch
			// ladders align across every row.
			col := func(w int, inner layout.Widget) layout.FlexChild {
				return layout.Rigid(func(gtx C) D {
					d := inner(gtx)
					d.Size.X = gtx.Dp(unit.Dp(w))
					return d
				})
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					gtx.Constraints.Min.X = gtx.Dp(110)
					return lotusui.Button(th, btn, s.name, lotusui.ButtonProps{Color: s.scale})(gtx)
				}),
				layout.Rigid(lotusui.HSpacer(lotusui.Space.MD)),
				col(96, lotusui.Badge(th, s.name, lotusui.BadgeProps{Color: s.scale})),
				col(48, lotusui.Avatar(th, lotusui.AvatarProps{Initials: s.name[:1], Color: s.scale})),
				layout.Rigid(lotusui.HSpacer(lotusui.Space.MD)),
				layout.Rigid(func(gtx C) D { return swatchRow(th, gtx, s.scale) }),
			)
		})
	}
	lotusui.Fill(gtx, th.Palette.Bg)
	return layout.UniformInset(lotusui.Space.XL).Layout(gtx, func(gtx C) D {
		return layout.Center.Layout(gtx, lotusui.VStack(lotusui.Space.LG, rows...))
	})
}

var showcaseColorBtns [5]widget.Clickable

// swatchRow paints the scale's ladder C100…C700 as rounded chips —
// the "one anchor in, the whole ladder out" visual.
func swatchRow(th *lotusui.Theme, gtx C, s lotusui.ColorScale) D {
	var chips []layout.FlexChild
	for _, c := range []color.NRGBA{s.C100, s.C200, s.C300, s.C400, s.C500, s.C600, s.C700} {
		c := c
		chips = append(chips, layout.Rigid(func(gtx C) D {
			d := gtx.Dp(24)
			sz := image.Pt(d, d)
			defer clip.UniformRRect(image.Rectangle{Max: sz}, gtx.Dp(6)).Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, c)
			return D{Size: sz}
		}), layout.Rigid(lotusui.HSpacer(6)))
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, chips...)
}

// showcaseDevicesDemo: the same components at desktop and phone
// widths, side by side — one codebase, two form factors. The desktop
// surface is chromeless with traffic lights ON the content (the
// seamless window, shown rather than told); the phone wears a slim
// bezel.
var dev struct {
	tabs          lotusui.Tabs
	invite, mSave widget.Clickable
	name          lotusui.Input
	notif         lotusui.Switch
	plan          lotusui.RadioGroup
	inited        bool
}

func showcaseDevicesDemo(th *lotusui.Theme, gtx C) D {
	if !dev.inited {
		dev.inited = true
		dev.tabs.Options = lotusui.TabOpts("Members", "Roles", "Billing")
		dev.name.Editor.SetText("Ada Lovelace")
		dev.notif.Value = true
		dev.plan.Options = lotusui.RadioOpts("Starter", "Pro", "Team")
		dev.plan.SetValue("Pro")
	}
	dev.tabs.Update(gtx)

	deviceFrame := func(w, h, bezel, radius int, content layout.Widget) layout.Widget {
		return func(gtx C) D {
			outer := image.Pt(gtx.Dp(unit.Dp(w)), gtx.Dp(unit.Dp(h)))
			r := gtx.Dp(unit.Dp(radius))
			// A soft drop shadow (the gallery can't reach the library's
			// unexported helper): translucent rings, biased down.
			for i := 3; i >= 1; i-- {
				grow := gtx.Dp(unit.Dp(i))
				sr := image.Rect(-grow, -grow+gtx.Dp(2), outer.X+grow, outer.Y+grow+gtx.Dp(2))
				paint.FillShape(gtx.Ops, color.NRGBA{A: uint8(16 - 4*i)}, clip.UniformRRect(sr, r+grow).Op(gtx.Ops))
			}
			defer clip.UniformRRect(image.Rectangle{Max: outer}, r).Push(gtx.Ops).Pop()
			if bezel > 0 {
				lotusui.Fill(gtx, th.Palette.BgInverted)
			}
			bz := gtx.Dp(unit.Dp(bezel))
			inner := image.Pt(outer.X-2*bz, outer.Y-2*bz)
			off := op.Offset(image.Pt(bz, bz)).Push(gtx.Ops)
			ir := r - bz
			if ir < 0 {
				ir = 0
			}
			cl := clip.UniformRRect(image.Rectangle{Max: inner}, ir).Push(gtx.Ops)
			cgtx := gtx
			cgtx.Constraints = layout.Constraints{Min: inner, Max: inner}
			lotusui.Fill(cgtx, th.Palette.Bg)
			content(cgtx)
			cl.Pop()
			off.Pop()
			return D{Size: outer}
		}
	}
	trafficLights := func(gtx C) D {
		cols := []color.NRGBA{
			{R: 0xFF, G: 0x5F, B: 0x57, A: 0xFF},
			{R: 0xFE, G: 0xBC, B: 0x2E, A: 0xFF},
			{R: 0x28, G: 0xC8, B: 0x40, A: 0xFF},
		}
		d := gtx.Dp(11)
		gap := gtx.Dp(7)
		for i, c := range cols {
			off := op.Offset(image.Pt(i*(d+gap), 0)).Push(gtx.Ops)
			func() {
				defer clip.UniformRRect(image.Rectangle{Max: image.Pt(d, d)}, d/2).Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, c)
			}()
			off.Pop()
		}
		return D{Size: image.Pt(3*d+2*gap, d)}
	}
	desktop := deviceFrame(520, 372, 0, 12, func(gtx C) D {
		return layout.UniformInset(lotusui.Space.LG).Layout(gtx, func(gtx C) D {
			return lotusui.VStack(th.Space.MD,
				trafficLights,
				func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
						layout.Rigid(lotusui.LabelTitle(th, "Team").Layout),
						layout.Rigid(lotusui.Button(th, &dev.invite, "Invite member", lotusui.ButtonProps{Size: lotusui.SizeSM})),
					)
				},
				func(gtx C) D { return dev.tabs.Layout(th, gtx) },
				func(gtx C) D {
					return lotusui.TableText(th, lotusui.TableProps{},
						[]string{"Member", "Role", "Status"},
						[][]string{
							{"Ada Lovelace", "Owner", "Active"},
							{"Edsger Dijkstra", "Editor", "Active"},
							{"Grace Hopper", "Viewer", "Invited"},
						})(gtx)
				},
			)(gtx)
		})
	})
	phone := deviceFrame(220, 372, 7, 28, func(gtx C) D {
		return layout.UniformInset(lotusui.Space.MD).Layout(gtx, func(gtx C) D {
			return lotusui.VStack(th.Space.MD,
				lotusui.LabelTitle(th, "Profile").Layout,
				lotusui.Field(th, lotusui.FieldProps{Label: "Name"}, func(gtx C) D {
					return dev.name.LayoutField(th, gtx, "")
				}),
				func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, lotusui.LabelBody(th, "Notifications").Layout),
						layout.Rigid(func(gtx C) D { return dev.notif.Layout(th, gtx) }),
					)
				},
				func(gtx C) D { return dev.plan.Layout(th, gtx) },
				lotusui.FullWidth(lotusui.Button(th, &dev.mSave, "Save", lotusui.ButtonProps{})),
			)(gtx)
		})
	})
	lotusui.Fill(gtx, th.Palette.Bg)
	return layout.UniformInset(lotusui.Space.XL).Layout(gtx, func(gtx C) D {
		return layout.Center.Layout(gtx, func(gtx C) D {
			return layout.Flex{Alignment: layout.Start}.Layout(gtx,
				layout.Rigid(desktop),
				layout.Rigid(lotusui.HSpacer(lotusui.Space.XL)),
				layout.Rigid(phone),
			)
		})
	})
}
