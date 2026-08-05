package main

// The demos for the second wave of components — one section per
// capability, indexes matching the docs pages' Demo refs.

import (
	"fmt"
	"image"
	"strconv"
	"time"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"

	lotusui "github.com/ikaito-com/lotusui"
)

// ---- accordion ----

var acc lotusui.Accordion

var accMulti = lotusui.Accordion{Multiple: true}
var accDis lotusui.Accordion
var accBorder = lotusui.Accordion{Bordered: true}

func accordionDemo(th *lotusui.Theme, gtx C) D {
	body := func(text string) layout.Widget {
		return func(gtx C) D {
			l := lotusui.LabelBody(th, text)
			l.Color = th.Palette.FgMuted
			return l.Layout(gtx)
		}
	}
	return card(th, gtx,
		section(th, "Accordion — one panel at a time", func(gtx C) D {
			return acc.Layout(th, gtx,
				lotusui.AccordionItem{Title: "Is it accessible?", Content: body("Yes. Keyboard dismissal and focus follow the same rules as every component.")},
				lotusui.AccordionItem{Title: "Is it styled?", Content: body("Yes. It comes with default styles that match the other components' aesthetic.")},
				lotusui.AccordionItem{Title: "Is it animated?", Content: body("Yes. It's animated by default on the shared clock — content never reflows mid-flight.")},
			)
		}),
		section(th, "Multiple — independent panels", func(gtx C) D {
			return accMulti.Layout(th, gtx,
				lotusui.AccordionItem{Title: "Notifications", Content: body("Choose how you want to be notified about activity: email, push, or in-app only.")},
				lotusui.AccordionItem{Title: "Privacy", Content: body("Control who can see your profile and activity.")},
				lotusui.AccordionItem{Title: "Billing", Content: body("Manage your payment methods and invoices.")},
			)
		}),
		section(th, "Disabled item", func(gtx C) D {
			return accDis.Layout(th, gtx,
				lotusui.AccordionItem{Title: "Can I access my account history?", Content: body("Yes, you can view your complete account history including all transactions, plan changes, and support tickets.")},
				lotusui.AccordionItem{Title: "Premium feature information", Disabled: true, Content: body("Upgrade your plan to access this content.")},
				lotusui.AccordionItem{Title: "How do I update my email address?", Content: body("You can update your email address in your account settings.")},
			)
		}),
		section(th, "Borders — the boxed look", func(gtx C) D {
			return accBorder.Layout(th, gtx,
				lotusui.AccordionItem{Title: "Billing", Content: body("Manage your payment methods and invoices.")},
				lotusui.AccordionItem{Title: "Team", Content: body("Invite and manage the people in your workspace.")},
			)
		}),
		section(th, "Card — inside a padded surface", func(gtx C) D {
			return lotusui.SurfaceCard(th, gtx, func(gtx C) D {
				return accCard.Layout(th, gtx,
					lotusui.AccordionItem{Title: "What is included?", Content: body("Every component, the theme engine, and the registry tooling.")},
					lotusui.AccordionItem{Title: "Can I change my plan later?", Content: body("Yes — upgrades apply immediately, downgrades at the next cycle.")},
				)
			})
		}),
	)
}

var accCard lotusui.Accordion

// ---- annotated-text ----

var (
	annAPI, annSLA lotusui.HoverCard
)

func annotatedTextDemo(th *lotusui.Theme, gtx C) D {
	annAPI.Width = 200
	annSLA.Width = 200
	annAPI.OpenDelay = 400 * time.Millisecond
	annSLA.OpenDelay = 400 * time.Millisecond
	annAPI.Side = lotusui.HoverCardTop
	annSLA.Side = lotusui.HoverCardTop
	terms := []lotusui.GlossaryTerm{
		{Term: "API", Tip: "Application Programming Interface"},
		{Term: "SLA", Tip: "Service Level Agreement"},
	}
	return card(th, gtx,
		section(th, "AnnotatedText — glossary terms open HoverCards", lotusui.AnnotatedText(th,
			"The API must meet the SLA under load.",
			terms, []*lotusui.HoverCard{&annAPI, &annSLA},
		)),
	)
}

// ---- alert ----

func alertDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "Alert", func(gtx C) D {
			return lotusui.Alert(th, lotusui.AlertProps{
				Title:       "Heads up!",
				Description: "You can add components to your app using the CLI.",
			})(gtx)
		}),
		section(th, "Destructive", func(gtx C) D {
			return lotusui.Alert(th, lotusui.AlertProps{
				Variant:     lotusui.AlertDestructive,
				Title:       "Error",
				Description: "Your session has expired. Please log in again.",
			})(gtx)
		}),
		section(th, "With action", func(gtx C) D {
			return lotusui.Alert(th, lotusui.AlertProps{
				Title:       "Dark mode is now available",
				Description: "Enable it under your profile settings to get started.",
				Action:      lotusui.Button(th, &alertEnable, "Enable", lotusui.ButtonProps{Size: lotusui.SizeXS}),
			})(gtx)
		}),
		section(th, "Colors — any scale, the pastel way", func(gtx C) D {
			return lotusui.Alert(th, lotusui.AlertProps{
				Color:       lotusui.Orange,
				Title:       "Your subscription will expire in 3 days.",
				Description: "Renew now to avoid service interruption or upgrade to a paid plan to continue using the service.",
			})(gtx)
		}),
	)
}

var alertEnable widget.Clickable

// ---- alert dialog ----

var ad struct {
	dlg                lotusui.AlertDialog
	open               widget.Clickable
	openDel, openSmall widget.Clickable
	mode               int // 0 default, 1 destructive, 2 small
	isOpen             bool
}

func alertDialogDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "AlertDialog — no backdrop dismissal, a decision required", func(gtx C) D {
			if ad.open.Clicked(gtx) && !ad.isOpen {
				ad.mode = 0
				ad.isOpen = true
				ownOverlay()
				ad.dlg.Appear()
			}
			return lotusui.Button(th, &ad.open, "Show Dialog", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})(gtx)
		}),
		section(th, "Destructive — a media medallion and a red exit", func(gtx C) D {
			if ad.openDel.Clicked(gtx) && !ad.isOpen {
				ad.mode = 1
				ad.isOpen = true
				ownOverlay()
				ad.dlg.Appear()
			}
			return lotusui.Button(th, &ad.openDel, "Delete Chat", lotusui.ButtonProps{Variant: lotusui.ButtonDestructive})(gtx)
		}),
		section(th, "Small", func(gtx C) D {
			if ad.openSmall.Clicked(gtx) && !ad.isOpen {
				ad.mode = 2
				ad.isOpen = true
				ownOverlay()
				ad.dlg.Appear()
			}
			return lotusui.Button(th, &ad.openSmall, "Show Dialog", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})(gtx)
		}),
	)
}

func alertDialogOverlay(th *lotusui.Theme, gtx C) D {
	if !ad.isOpen {
		return D{}
	}
	if ad.dlg.Confirmed(gtx) || ad.dlg.Cancelled(gtx) {
		ad.isOpen = false
		return D{}
	}
	switch ad.mode {
	case 1:
		return ad.dlg.Layout(th, gtx, lotusui.AlertDialogProps{
			Size:        lotusui.SizeSM,
			Title:       "Delete chat?",
			Description: "This will permanently delete this chat conversation and any memories saved during this chat.",
			Action:      "Delete",
			Destructive: true,
			Media: func(gtx C) D {
				d := gtx.Dp(40)
				sz := image.Pt(d, d)
				defer clip.UniformRRect(image.Rectangle{Max: sz}, gtx.Dp(10)).Push(gtx.Ops).Pop()
				lotusui.Fill(gtx, th.Palette.DangerBg)
				cgtx := gtx
				cgtx.Constraints = layout.Constraints{Min: sz, Max: sz}
				layout.Center.Layout(cgtx, lotusui.SVGIcon(lotusui.IconRefuse, 20, th.Palette.Danger))
				return D{Size: sz}
			},
		})
	case 2:
		return ad.dlg.Layout(th, gtx, lotusui.AlertDialogProps{
			Size:        lotusui.SizeSM,
			Title:       "Allow accessory to connect?",
			Description: "Do you want to allow the USB accessory to connect to this device?",
			Cancel:      "Don't allow",
			Action:      "Allow",
		})
	}
	return ad.dlg.Layout(th, gtx, lotusui.AlertDialogProps{
		Title:       "Are you absolutely sure?",
		Description: "This action cannot be undone. This will permanently delete your account and remove your data from our servers.",
	})
}

// ---- avatar ----

func avatarDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "Avatar — initials, fallback glyph", lotusui.HStack(lotusui.Space.SM,
			lotusui.Avatar(th, lotusui.AvatarProps{Initials: "CN"}),
			lotusui.Avatar(th, lotusui.AvatarProps{Initials: "ER", Color: lotusui.Teal}),
			lotusui.Avatar(th, lotusui.AvatarProps{Initials: "LR", Color: lotusui.Pink}),
			lotusui.Avatar(th, lotusui.AvatarProps{}),
		)),
		section(th, "Sizes — 2XS to 2XL", lotusui.HStack(lotusui.Space.SM,
			lotusui.Avatar(th, lotusui.AvatarProps{Initials: "A", Size: lotusui.Size2XS}),
			lotusui.Avatar(th, lotusui.AvatarProps{Initials: "A", Size: lotusui.SizeXS}),
			lotusui.Avatar(th, lotusui.AvatarProps{Initials: "A", Size: lotusui.SizeSM}),
			lotusui.Avatar(th, lotusui.AvatarProps{Initials: "A"}),
			lotusui.Avatar(th, lotusui.AvatarProps{Initials: "A", Size: lotusui.SizeLG}),
			lotusui.Avatar(th, lotusui.AvatarProps{Initials: "A", Size: lotusui.SizeXL}),
			lotusui.Avatar(th, lotusui.AvatarProps{Initials: "A", Size: lotusui.Size2XL}),
		)),
		section(th, "Group — ringed overlap", func(gtx C) D {
			return lotusui.AvatarGroup(th, lotusui.AvatarGroupProps{},
				lotusui.AvatarProps{Initials: "CN"},
				lotusui.AvatarProps{Initials: "ER", Color: lotusui.Teal},
				lotusui.AvatarProps{Initials: "LR", Color: lotusui.Pink},
			)(gtx)
		}),
		section(th, "Badge — status on the rim", lotusui.HStack(lotusui.Space.SM,
			lotusui.Avatar(th, lotusui.AvatarProps{Initials: "CN",
				Badge: &lotusui.AvatarBadge{Color: lotusui.Green}}),
			lotusui.Avatar(th, lotusui.AvatarProps{Initials: "PP", Color: lotusui.Purple,
				Badge: &lotusui.AvatarBadge{Icon: lotusui.IconPlus}}),
		)),
		section(th, "Group with count", lotusui.HStack(lotusui.Space.LG,
			lotusui.AvatarGroup(th, lotusui.AvatarGroupProps{Count: "+3"},
				lotusui.AvatarProps{Initials: "CN"},
				lotusui.AvatarProps{Initials: "LR", Color: lotusui.Teal},
				lotusui.AvatarProps{Initials: "ER", Color: lotusui.Pink},
			),
			lotusui.AvatarGroup(th, lotusui.AvatarGroupProps{CountIcon: lotusui.IconPlus},
				lotusui.AvatarProps{Initials: "CN"},
				lotusui.AvatarProps{Initials: "LR", Color: lotusui.Teal},
				lotusui.AvatarProps{Initials: "ER", Color: lotusui.Pink},
			),
		)),
	)
}

// ---- breadcrumb ----

var bcBtns = []*widget.Clickable{new(widget.Clickable), new(widget.Clickable)}
var bcHome, bcComp widget.Clickable
var bcEllipsisMenu lotusui.DropdownMenuTrigger // hand-composed … menu (Usage)
var bcMenu lotusui.DropdownMenuTrigger         // "Components ▾" dropdown
var bcMenuItems struct{ docs, themes, github widget.Clickable }
var bcDropItems struct{ docs, themes, github widget.Clickable }
var bcNav lotusui.BreadcrumbNav
var bcNavBtns struct{ home, docs, build, fetch widget.Clickable }

func breadcrumbDemo(th *lotusui.Theme, gtx C) D {
	bcEllipsisMenu.Variant = lotusui.ButtonGhost
	bcEllipsisMenu.Icon = lotusui.IconMoreHorizontal
	bcEllipsisMenu.Size = lotusui.SizeSM
	bcEllipsisMenu.Width = 160
	bcEllipsisMenu.Align = lotusui.PopoverStart
	bcMenu.Variant = lotusui.ButtonGhost
	bcMenu.Width = 160
	bcMenu.Align = lotusui.PopoverStart
	sep := func(gtx C) D { return lotusui.BreadcrumbSep(th, lotusui.IconDot)(gtx) }
	chev := func(gtx C) D { return lotusui.BreadcrumbSep(th, "")(gtx) }
	return card(th, gtx,
		section(th, "Breadcrumb — ellipsis opens the collapsed trail", func(gtx C) D {
			// Mirrors shadcn breadcrumb-demo.tsx: Home / …menu / Components / Breadcrumb
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(lotusui.BreadcrumbLink(th, &bcHome, "Home")),
				layout.Rigid(chev),
				layout.Rigid(func(gtx C) D {
					return bcEllipsisMenu.Layout(th, gtx, "",
						lotusui.DropdownMenuItem(th, &bcMenuItems.docs, "Documentation", false),
						lotusui.DropdownMenuItem(th, &bcMenuItems.themes, "Themes", false),
						lotusui.DropdownMenuItem(th, &bcMenuItems.github, "GitHub", false),
					)
				}),
				layout.Rigid(chev),
				layout.Rigid(lotusui.BreadcrumbLink(th, &bcComp, "Components")),
				layout.Rigid(chev),
				layout.Rigid(lotusui.BreadcrumbPage(th, "Breadcrumb")),
			)
		}),
		section(th, "Custom separator — composed from the pieces", func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(lotusui.BreadcrumbLink(th, &bcHome, "Home")),
				layout.Rigid(sep),
				layout.Rigid(lotusui.BreadcrumbLink(th, &bcComp, "Components")),
				layout.Rigid(sep),
				layout.Rigid(lotusui.BreadcrumbPage(th, "Breadcrumb")),
			)
		}),
		section(th, "Dropdown — a trail label opens a menu", func(gtx C) D {
			// Mirrors shadcn breadcrumb-dropdown.tsx: Home / Components▾ / Breadcrumb
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(lotusui.BreadcrumbLink(th, &bcHome, "Home")),
				layout.Rigid(sep),
				layout.Rigid(func(gtx C) D {
					return bcMenu.Layout(th, gtx, "Components",
						lotusui.DropdownMenuItem(th, &bcDropItems.docs, "Documentation", false),
						lotusui.DropdownMenuItem(th, &bcDropItems.themes, "Themes", false),
						lotusui.DropdownMenuItem(th, &bcDropItems.github, "GitHub", false),
					)
				}),
				layout.Rigid(sep),
				layout.Rigid(lotusui.BreadcrumbPage(th, "Breadcrumb")),
			)
		}),
		section(th, "Collapsed — bare BreadcrumbEllipsis", func(gtx C) D {
			// Mirrors shadcn breadcrumb-ellipsis.tsx (glyph only).
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(lotusui.BreadcrumbLink(th, &bcHome, "Home")),
				layout.Rigid(chev),
				layout.Rigid(lotusui.BreadcrumbEllipsis(th)),
				layout.Rigid(chev),
				layout.Rigid(lotusui.BreadcrumbLink(th, &bcComp, "Components")),
				layout.Rigid(chev),
				layout.Rigid(lotusui.BreadcrumbPage(th, "Breadcrumb")),
			)
		}),
		section(th, "Responsive — BreadcrumbNav collapses the middle", func(gtx C) D {
			// Mirrors shadcn breadcrumb-responsive.tsx: ITEMS_TO_DISPLAY=3,
			// ellipsis menu holds the middle, below md shows first+last + truncate.
			return bcNav.Layout(th, gtx,
				lotusui.BreadcrumbSegLink(&bcNavBtns.home, "Home"),
				lotusui.BreadcrumbSegLink(&bcNavBtns.docs, "Documentation"),
				lotusui.BreadcrumbSegLink(&bcNavBtns.build, "Building Your Application"),
				lotusui.BreadcrumbSegLink(&bcNavBtns.fetch, "Data Fetching"),
				lotusui.BreadcrumbSegOf("Caching and Revalidating"),
			)
		}),
	)
}

// ---- button group ----

var bg struct {
	back, archive, report, snooze widget.Clickable
	more                          lotusui.DropdownMenuTrigger
	moreA, moreB, moreC           widget.Clickable
	vPlus, vMinus                 widget.Clickable
	szBtns                        [9]widget.Clickable
	copyBtn, pasteBtn             widget.Clickable
	splitMain, splitPlus          widget.Clickable
	search                        lotusui.Input
	searchBtn                     widget.Clickable
}

func buttonGroupDemo(th *lotusui.Theme, gtx C) D {
	outline := func(btn *widget.Clickable, label string, opts lotusui.ButtonProps) lotusui.ButtonGroupItem {
		opts.Variant = lotusui.ButtonOutline
		return lotusui.ButtonGroupItem{Btn: btn, Label: label, Props: opts}
	}
	return card(th, gtx,
		section(th, "ButtonGroup — attached, one shared border", func(gtx C) D {
			bg.more.Variant = lotusui.ButtonGhost
			return lotusui.HStack(th.Space.SM,
				lotusui.ButtonGroup(th, lotusui.ButtonGroupProps{},
					outline(&bg.back, "", lotusui.ButtonProps{IconStart: lotusui.IconArrowLeft}),
				),
				lotusui.ButtonGroup(th, lotusui.ButtonGroupProps{},
					outline(&bg.archive, "Archive", lotusui.ButtonProps{IconStart: lotusui.IconFile}),
					outline(&bg.report, "Report", lotusui.ButtonProps{IconStart: lotusui.IconWarning}),
					outline(&bg.snooze, "Snooze", lotusui.ButtonProps{IconStart: lotusui.IconClock}),
				),
				func(gtx C) D {
					bg.more.Width = 180
					bg.more.Align = lotusui.PopoverEnd
					return bg.more.Layout(th, gtx, "⋯",
						lotusui.DropdownMenuItem(th, &bg.moreA, "Mark as read", false),
						lotusui.DropdownMenuItem(th, &bg.moreB, "Add label", false),
						lotusui.DropdownMenuSeparator(th),
						lotusui.DropdownMenuItem(th, &bg.moreC, "Trash", true),
					)
				},
			)(gtx)
		}),
		section(th, "Orientation — vertical", func(gtx C) D {
			return lotusui.ButtonGroup(th, lotusui.ButtonGroupProps{Vertical: true},
				outline(&bg.vPlus, "", lotusui.ButtonProps{IconStart: lotusui.IconPlus}),
				outline(&bg.vMinus, "", lotusui.ButtonProps{IconStart: lotusui.IconMinus}),
			)(gtx)
		}),
		section(th, "Sizes", func(gtx C) D {
			rowAt := func(off int, sz lotusui.Size) layout.Widget {
				return lotusui.ButtonGroup(th, lotusui.ButtonGroupProps{},
					outline(&bg.szBtns[off], "Small", lotusui.ButtonProps{Size: sz}),
					outline(&bg.szBtns[off+1], "Button", lotusui.ButtonProps{Size: sz}),
					outline(&bg.szBtns[off+2], "", lotusui.ButtonProps{Size: sz, IconStart: lotusui.IconPlus}),
				)
			}
			return lotusui.VStack(th.Space.MD,
				rowAt(0, lotusui.SizeSM),
				rowAt(3, lotusui.SizeMD),
				rowAt(6, lotusui.SizeLG),
			)(gtx)
		}),
		section(th, "Separator — the visible seam", func(gtx C) D {
			sec := func(btn *widget.Clickable, label string) lotusui.ButtonGroupItem {
				return lotusui.ButtonGroupItem{Btn: btn, Label: label,
					Props: lotusui.ButtonProps{Variant: lotusui.ButtonSecondary, Size: lotusui.SizeSM}}
			}
			return lotusui.ButtonGroup(th, lotusui.ButtonGroupProps{},
				sec(&bg.copyBtn, "Copy"),
				lotusui.ButtonGroupSeparator(),
				sec(&bg.pasteBtn, "Paste"),
			)(gtx)
		}),
		section(th, "Split", func(gtx C) D {
			return lotusui.ButtonGroup(th, lotusui.ButtonGroupProps{},
				lotusui.ButtonGroupItem{Btn: &bg.splitMain, Label: "Button",
					Props: lotusui.ButtonProps{Variant: lotusui.ButtonSecondary}},
				lotusui.ButtonGroupSeparator(),
				lotusui.ButtonGroupItem{Btn: &bg.splitPlus,
					Props: lotusui.ButtonProps{Variant: lotusui.ButtonSecondary, IconStart: lotusui.IconPlus}},
			)(gtx)
		}),
		section(th, "With input", func(gtx C) D {
			return lotusui.ButtonGroup(th, lotusui.ButtonGroupProps{},
				lotusui.ButtonGroupInput(&bg.search, "Search...", 1),
				outline(&bg.searchBtn, "", lotusui.ButtonProps{IconStart: lotusui.IconSearch}),
			)(gtx)
		}),
	)
}

// ---- input otp ----

var otp struct {
	demo, sep, four lotusui.InputOTP
	digits, off     lotusui.InputOTP
	bad             lotusui.InputOTP
	inited          bool
}

func inputOTPDemo(th *lotusui.Theme, gtx C) D {
	if !otp.inited {
		otp.inited = true
		otp.demo.SetValue("123456")
		otp.sep.Groups = []int{2, 2, 2}
		otp.four.Length = 4
		otp.digits.Filter = "0123456789"
		otp.off.Disabled = true
		otp.off.SetValue("123")
		otp.bad.Invalid = true
		otp.bad.SetValue("123456")
	}
	return card(th, gtx,
		section(th, "InputOTP — click the row and type or paste", func(gtx C) D {
			return otp.demo.Layout(th, gtx)
		}),
		section(th, "Separator — grouped digits", func(gtx C) D {
			return otp.sep.Layout(th, gtx)
		}),
		section(th, "Four digits", func(gtx C) D {
			return otp.four.Layout(th, gtx)
		}),
		section(th, "Pattern — digits only", func(gtx C) D {
			return lotusui.Field(th, lotusui.FieldProps{Label: "Digits Only"}, func(gtx C) D {
				return otp.digits.Layout(th, gtx)
			})(gtx)
		}),
		section(th, "Disabled", func(gtx C) D {
			return otp.off.Layout(th, gtx)
		}),
		section(th, "Invalid", func(gtx C) D {
			return lotusui.Field(th, lotusui.FieldProps{Error: "The one-time password is incorrect."}, func(gtx C) D {
				return otp.bad.Layout(th, gtx)
			})(gtx)
		}),
	)
}

// ---- pagination ----

var pg = lotusui.Pagination{Page: 5}
var pgSimple = lotusui.Pagination{Page: 2, Simple: true}
var pgTools = lotusui.Pagination{Page: 3}
var pgRows = lotusui.Select{Options: lotusui.SelectOpts("10", "25", "50", "100")}

func paginationDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "Pagination — long ranges elide around the page", func(gtx C) D {
			return pg.Layout(th, gtx, 12)
		}),
		section(th, "Simple — numbered links only", func(gtx C) D {
			return pgSimple.Layout(th, gtx, 5)
		}),
		section(th, "Rows per page — a toolbar composition", func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(lotusui.LabelBody(th, "Rows per page").Layout),
						layout.Rigid(lotusui.HSpacer(th.Space.SM)),
						layout.Rigid(func(gtx C) D {
							gtx.Constraints.Max.X = gtx.Dp(90)
							return pgRows.Layout(th, gtx, "")
						}),
					)
				}),
				layout.Rigid(func(gtx C) D { return pgTools.Layout(th, gtx, 9) }),
			)
		}),
	)
}

// ---- hover-card ----

var (
	hcUsage, hcDelay, hcPos, hcBasic lotusui.HoverCard
	hcSides                          [4]lotusui.HoverCard
	hcBtnUsage, hcBtnDelay           widget.Clickable
	hcBtnPos, hcBtnBasic             widget.Clickable
	hcBtnSides                       [4]widget.Clickable
)

func hoverCardBody(th *lotusui.Theme) layout.Widget {
	return lotusui.VStack(th.Space.XS,
		func(gtx C) D {
			l := lotusui.LabelBody(th, "@nextjs")
			l.Font.Weight = font.SemiBold
			return l.Layout(gtx)
		},
		func(gtx C) D {
			l := lotusui.LabelMeta(th, "The React Framework – created and maintained by @vercel.")
			return l.Layout(gtx)
		},
		func(gtx C) D {
			return lotusui.HStack(lotusui.Space.XS,
				lotusui.SVGIcon(lotusui.IconClock, 14, th.Palette.FgSubtle),
				func(gtx C) D {
					l := lotusui.LabelCaption(th, "Joined December 2021")
					l.Color = th.Palette.FgMuted
					return l.Layout(gtx)
				},
			)(gtx)
		},
	)
}

func hoverCardDemo(th *lotusui.Theme, gtx C) D {
	hcDelay.OpenDelay = 100 * time.Millisecond
	hcDelay.CloseDelay = 200 * time.Millisecond
	hcPos.Side = lotusui.HoverCardTop
	hcPos.Align = lotusui.PopoverStart
	sides := []lotusui.HoverCardSide{lotusui.HoverCardLeft, lotusui.HoverCardTop, lotusui.HoverCardBottom, lotusui.HoverCardRight}
	labels := []string{"left", "top", "bottom", "right"}
	for i := range hcSides {
		hcSides[i].Side = sides[i]
		hcSides[i].OpenDelay = 100 * time.Millisecond
		hcSides[i].CloseDelay = 100 * time.Millisecond
	}
	hcUsage.OpenDelay = 100 * time.Millisecond
	hcBasic.OpenDelay = 100 * time.Millisecond
	return card(th, gtx,
		section(th, "HoverCard — rest the pointer on the trigger", func(gtx C) D {
			return hcUsage.Layout(th, gtx, hoverCardBody(th),
				lotusui.Button(th, &hcBtnUsage, "Hover", lotusui.ButtonProps{Variant: lotusui.ButtonLink}))
		}),
		section(th, "Trigger delays — OpenDelay / CloseDelay", func(gtx C) D {
			return hcDelay.Layout(th, gtx, hoverCardBody(th),
				lotusui.Button(th, &hcBtnDelay, "Hover", lotusui.ButtonProps{Variant: lotusui.ButtonLink}))
		}),
		section(th, "Positioning — Side + Align", func(gtx C) D {
			return hcPos.Layout(th, gtx, hoverCardBody(th),
				lotusui.Button(th, &hcBtnPos, "Hover", lotusui.ButtonProps{Variant: lotusui.ButtonLink}))
		}),
		section(th, "Basic — @nextjs profile preview", func(gtx C) D {
			return hcBasic.Layout(th, gtx,
				lotusui.HStack(th.Space.MD,
					lotusui.Avatar(th, lotusui.AvatarProps{Initials: "VC", Size: lotusui.SizeLG}),
					hoverCardBody(th),
				),
				lotusui.Button(th, &hcBtnBasic, "@nextjs", lotusui.ButtonProps{Variant: lotusui.ButtonLink}),
			)
		}),
		section(th, "Sides — left / top / bottom / right", func(gtx C) D {
			var row []layout.FlexChild
			for i := range labels {
				i := i
				row = append(row, layout.Rigid(func(gtx C) D {
					return hcSides[i].Layout(th, gtx, hoverCardBody(th),
						lotusui.Button(th, &hcBtnSides[i], labels[i], lotusui.ButtonProps{Variant: lotusui.ButtonOutline}))
				}))
				if i < len(labels)-1 {
					row = append(row, layout.Rigid(lotusui.HSpacer(th.Space.SM)))
				}
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
		}),
	)
}

// ---- popover ----

var pop struct {
	p       lotusui.Popover
	trigger widget.Clickable
}

var popAligns struct {
	p    [3]lotusui.Popover
	btns [3]widget.Clickable
}

var popW, popH lotusui.Input

func popoverDemo(th *lotusui.Theme, gtx C) D {
	if pop.trigger.Clicked(gtx) {
		pop.p.Open = !pop.p.Open
	}
	return card(th, gtx,
		section(th, "Popover — arbitrary content on the floating layer", func(gtx C) D {
			dims := lotusui.Button(th, &pop.trigger, "Open popover", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})(gtx)
			pop.p.Width = 260
			pop.p.Layout(th, gtx, dims.Size, lotusui.VStack(th.Space.SM,
				lotusui.LabelCardTitle(th, "Dimensions").Layout,
				func(gtx C) D {
					l := lotusui.LabelCaption(th, "Set the dimensions for the layer.")
					l.Color = th.Palette.FgMuted
					return l.Layout(gtx)
				},
				lotusui.Field(th, lotusui.FieldProps{Label: "Width"}, func(gtx C) D {
					return popW.LayoutField(th, gtx, "100%")
				}),
				lotusui.Field(th, lotusui.FieldProps{Label: "Height"}, func(gtx C) D {
					return popH.LayoutField(th, gtx, "25px")
				}),
			))
			return dims
		}),
		section(th, "Alignments — start, center, end against the anchor", func(gtx C) D {
			labels := []string{"Start", "Center", "End"}
			aligns := []lotusui.PopoverAlign{lotusui.PopoverStart, lotusui.PopoverCenter, lotusui.PopoverEnd}
			texts := []string{"Aligned to start", "Aligned to center", "Aligned to end"}
			var row []layout.FlexChild
			for i := range labels {
				i := i
				if popAligns.btns[i].Clicked(gtx) {
					popAligns.p[i].Open = !popAligns.p[i].Open
				}
				if i > 0 {
					row = append(row, layout.Rigid(lotusui.HSpacer(lotusui.Space.LG)))
				}
				row = append(row, layout.Rigid(func(gtx C) D {
					dims := lotusui.Button(th, &popAligns.btns[i], labels[i],
						lotusui.ButtonProps{Variant: lotusui.ButtonOutline, Size: lotusui.SizeSM})(gtx)
					popAligns.p[i].Width = 160
					popAligns.p[i].Align = aligns[i]
					popAligns.p[i].Layout(th, gtx, dims.Size, func(gtx C) D {
						return lotusui.LabelBody(th, texts[i]).Layout(gtx)
					})
					return dims
				}))
			}
			return layout.Flex{}.Layout(gtx, row...)
		}),
	)
}

// ---- progress ----

func progressDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "Progress", lotusui.VStack(th.Space.MD,
			lotusui.Progress(th, 0.25),
			lotusui.Progress(th, 0.66),
			lotusui.Progress(th, 1),
		)),
		section(th, "Indeterminate — a sweeping segment", func(gtx C) D {
			return lotusui.Progress(th, -1)(gtx)
		}),
		section(th, "Label — value beside the bar", func(gtx C) D {
			return lotusui.VStack(th.Space.SM,
				func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
						layout.Rigid(lotusui.LabelBody(th, "Uploading…").Layout),
						layout.Rigid(func(gtx C) D {
							l := lotusui.LabelCaption(th, "66%")
							l.Color = th.Palette.FgSubtle
							return l.Layout(gtx)
						}),
					)
				},
				lotusui.Progress(th, 0.66),
			)(gtx)
		}),
	)
}

// ---- radio group ----

var rg = lotusui.RadioGroup{Options: lotusui.RadioOpts("Default", "Comfortable", "Compact")}
var rgDis = lotusui.RadioGroup{Options: []lotusui.RadioOption{
	{Label: "Starter", Value: "starter"},
	{Label: "Pro (contact sales)", Value: "pro", Disabled: true},
	{Label: "Enterprise", Value: "enterprise"},
}}
var rgDesc = lotusui.RadioGroup{Options: []lotusui.RadioOption{
	{Label: "Default", Value: "default", Description: "Standard spacing for most use cases."},
	{Label: "Comfortable", Value: "comfortable", Description: "More space between elements."},
	{Label: "Compact", Value: "compact", Description: "Dense layout for scanning lots of data."},
}}
var rgBad = lotusui.RadioGroup{Invalid: true,
	Options: lotusui.RadioOpts("Email only", "SMS only", "Both")}

func radioDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "RadioGroup — one selected, always", func(gtx C) D {
			return rg.Layout(th, gtx)
		}),
		section(th, "Disabled option", func(gtx C) D {
			return rgDis.Layout(th, gtx)
		}),
		section(th, "Description", func(gtx C) D {
			return rgDesc.Layout(th, gtx)
		}),
		section(th, "Invalid", func(gtx C) D {
			return lotusui.Field(th, lotusui.FieldProps{Label: "Notification Preferences", Error: "choose how you want to receive notifications"}, func(gtx C) D {
				return rgBad.Layout(th, gtx)
			})(gtx)
		}),
	)
}

func init() { rgDesc.SetValue("comfortable") }

// ---- separator ----

func separatorDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "Separator — horizontal and vertical", func(gtx C) D {
			return lotusui.VStack(th.Space.MD,
				lotusui.LabelBody(th, "An open-source Go design system — desktop, mobile, and the web.").Layout,
				lotusui.Separator(th),
				func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(lotusui.LabelCaption(th, "Docs").Layout),
						layout.Rigid(lotusui.HSpacer(th.Space.SM)),
						layout.Rigid(lotusui.SeparatorVertical(th)),
						layout.Rigid(lotusui.HSpacer(th.Space.SM)),
						layout.Rigid(lotusui.LabelCaption(th, "Source").Layout),
					)
				},
			)(gtx)
		}),
		section(th, "List — items divided inline", func(gtx C) D {
			item := func(t string) layout.FlexChild {
				return layout.Rigid(lotusui.LabelBody(th, t).Layout)
			}
			div := func() []layout.FlexChild {
				return []layout.FlexChild{
					layout.Rigid(lotusui.HSpacer(th.Space.SM)),
					layout.Rigid(lotusui.SeparatorVertical(th)),
					layout.Rigid(lotusui.HSpacer(th.Space.SM)),
				}
			}
			row := []layout.FlexChild{item("Blog")}
			row = append(row, div()...)
			row = append(row, item("Docs"))
			row = append(row, div()...)
			row = append(row, item("Source"))
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
		}),
	)
}

// ---- skeleton ----

func skeletonDemo(th *lotusui.Theme, gtx C) D {
	line := func(w, h int) layout.Widget { return lotusui.Skeleton(th, unitDp(w), unitDp(h)) }
	return card(th, gtx,
		section(th, "Skeleton", func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(lotusui.SkeletonCircle(th, 48)),
				layout.Rigid(lotusui.HSpacer(lotusui.Space.MD)),
				layout.Rigid(lotusui.VStack(lotusui.Space.SM,
					line(250, 16),
					line(200, 16),
				)),
			)
		}),
		section(th, "Avatar", func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(lotusui.SkeletonCircle(th, 40)),
				layout.Rigid(lotusui.HSpacer(lotusui.Space.MD)),
				layout.Rigid(lotusui.VStack(lotusui.Space.SM,
					line(150, 16),
					line(100, 16),
				)),
			)
		}),
		section(th, "Card", func(gtx C) D {
			gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(320))
			return lotusui.Card(th, lotusui.CardProps{}, lotusui.VStack(lotusui.Space.SM,
				line(180, 16),
				line(140, 16),
				func(gtx C) D {
					w := gtx.Constraints.Max.X
					return lotusui.Skeleton(th, 0, unitDp(w*9/16/2))(gtx) // 16:9 media block
				},
			))(gtx)
		}),
		section(th, "Form", func(gtx C) D {
			gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(320))
			return lotusui.VStack(lotusui.Space.LG,
				lotusui.VStack(lotusui.Space.SM, line(80, 16), lotusui.Skeleton(th, 0, 32)),
				lotusui.VStack(lotusui.Space.SM, line(96, 16), lotusui.Skeleton(th, 0, 32)),
				line(96, 32),
			)(gtx)
		}),
		section(th, "Table", func(gtx C) D {
			gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(380))
			var rows []layout.Widget
			for i := 0; i < 5; i++ {
				rows = append(rows, func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, lotusui.Skeleton(th, 0, 16)),
						layout.Rigid(lotusui.HSpacer(lotusui.Space.MD)),
						layout.Rigid(line(96, 16)),
						layout.Rigid(lotusui.HSpacer(lotusui.Space.MD)),
						layout.Rigid(line(80, 16)),
					)
				})
			}
			return lotusui.VStack(lotusui.Space.SM, rows...)(gtx)
		}),
		section(th, "Text", func(gtx C) D {
			gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(320))
			return lotusui.VStack(lotusui.Space.SM,
				lotusui.Skeleton(th, 0, 16),
				lotusui.Skeleton(th, 0, 16),
				func(gtx C) D {
					gtx.Constraints.Max.X = gtx.Constraints.Max.X * 3 / 4
					return lotusui.Skeleton(th, 0, 16)(gtx)
				},
			)(gtx)
		}),
	)
}

// unitDp converts a plain int for the skeleton shorthand above.
func unitDp(v int) unit.Dp { return unit.Dp(v) }

// ---- slider ----

var sld = lotusui.Slider{Value: 0.4}
var sldStep = lotusui.Slider{Value: 0.5, Step: 0.25}
var sldOff = lotusui.Slider{Value: 0.6, Disabled: true}

func sliderDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "Slider — drag the thumb or press the track", func(gtx C) D {
			return lotusui.VStack(th.Space.SM,
				func(gtx C) D { return sld.Layout(th, gtx) },
				func(gtx C) D {
					l := lotusui.LabelCaption(th, fmt.Sprintf("%.0f%%", sld.Value*100))
					l.Color = th.Palette.FgSubtle
					return l.Layout(gtx)
				},
			)(gtx)
		}),
		section(th, "Step — snapped to quarters", func(gtx C) D {
			return sldStep.Layout(th, gtx)
		}),
		section(th, "Disabled", func(gtx C) D {
			return sldOff.Layout(th, gtx)
		}),
		section(th, "Range — two thumbs, the fill between them", func(gtx C) D {
			return sldRange.Layout(th, gtx)
		}),
		section(th, "Multiple — one thumb per value", func(gtx C) D {
			return sldMulti.Layout(th, gtx)
		}),
		section(th, "Vertical — the value grows upward", func(gtx C) D {
			return lotusui.HStack(lotusui.Space.LG,
				func(gtx C) D {
					gtx.Constraints.Max.Y = gtx.Dp(160)
					return sldV1.Layout(th, gtx)
				},
				func(gtx C) D {
					gtx.Constraints.Max.Y = gtx.Dp(160)
					return sldV2.Layout(th, gtx)
				},
			)(gtx)
		}),
	)
}

var (
	sldRange = lotusui.Slider{Values: []float32{0.25, 0.5}, Step: 0.05}
	sldMulti = lotusui.Slider{Values: []float32{0.1, 0.2, 0.7}, Step: 0.1}
	sldV1    = lotusui.Slider{Value: 0.5, Vertical: true}
	sldV2    = lotusui.Slider{Value: 0.25, Vertical: true}
)

// ---- spinner ----

func spinnerDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "Spinner", lotusui.HStack(lotusui.Space.MD,
			lotusui.Spinner(th, 16),
			lotusui.Spinner(th, 24),
			lotusui.SpinnerTint(th, 24, th.Palette.BrandFg),
		)),
		section(th, "Badge — syncing states", lotusui.HStack(lotusui.Space.MD,
			lotusui.Badge(th, "Syncing", lotusui.BadgeProps{
				Start: lotusui.SpinnerTint(th, 12, th.Palette.Accent().OnSubtle)}),
			lotusui.Badge(th, "Updating", lotusui.BadgeProps{Variant: lotusui.BadgeSecondary,
				Start: lotusui.Spinner(th, 12)}),
			lotusui.Badge(th, "Processing", lotusui.BadgeProps{Variant: lotusui.BadgeOutline,
				Start: lotusui.Spinner(th, 12)}),
		)),
	)
}

// ---- table ----

func tableDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "Table — muted header, hairline rows", func(gtx C) D {
			return lotusui.TableText(th, lotusui.TableProps{Caption: "A list of recent invoices.", Footer: []string{"Total", "", "", "$750.00"}},
				[]string{"Invoice", "Status", "Method", "Amount"},
				[][]string{
					{"INV-001", "Paid", "Credit card", "$250.00"},
					{"INV-002", "Pending", "Transfer", "$150.00"},
					{"INV-003", "Unpaid", "Credit card", "$350.00"},
				})(gtx)
		}),
		section(th, "Actions — a menu per row", func(gtx C) D {
			cellText := func(txt string) layout.Widget {
				return func(gtx C) D {
					l := lotusui.LabelBody(th, txt)
					l.Color = th.Palette.Fg
					l.MaxLines = 1
					return l.Layout(gtx)
				}
			}
			row := func(i int, name, price string) []layout.Widget {
				return []layout.Widget{cellText(name), cellText(price), func(gtx C) D {
					return lotusui.RightAligned(func(gtx C) D {
						tblMenus[i].Variant = lotusui.ButtonGhost
						tblMenus[i].Align = lotusui.PopoverEnd
						tblMenus[i].Width = 160
						return tblMenus[i].Layout(th, gtx, "⋯",
							lotusui.DropdownMenuItem(th, &tblEdit[i], "Edit", false),
							lotusui.DropdownMenuItem(th, &tblDup[i], "Duplicate", false),
							lotusui.DropdownMenuSeparator(th),
							lotusui.DropdownMenuItem(th, &tblDel[i], "Delete", true),
						)
					})(gtx)
				}}
			}
			return lotusui.Table(th, lotusui.TableProps{},
				[]string{"Product", "Price", "Actions"},
				[][]layout.Widget{
					row(0, "Wireless Mouse", "$29.99"),
					row(1, "Mechanical Keyboard", "$129.99"),
				})(gtx)
		}),
	)
}

var (
	tblMenus                [2]lotusui.DropdownMenuTrigger
	tblEdit, tblDup, tblDel [2]widget.Clickable
)

// ---- textarea ----

var ta, taErr, taOff, taSend lotusui.Textarea
var taSendBtn widget.Clickable

func textareaDemo(th *lotusui.Theme, gtx C) D {
	if taErr.Error == "" {
		taErr.Error = "your message is required"
	}
	return card(th, gtx,
		section(th, "Textarea", func(gtx C) D {
			return ta.LayoutField(th, gtx, "Type your message here…")
		}),
		section(th, "Invalid — with a Field label", func(gtx C) D {
			return taErr.Layout(th, gtx, "Message", "Tell us what happened")
		}),
		section(th, "Disabled", func(gtx C) D {
			taOff.Disabled = true
			return lotusui.Field(th, lotusui.FieldProps{Label: "Message"}, func(gtx C) D {
				return taOff.LayoutField(th, gtx, "Type your message here.")
			})(gtx)
		}),
		section(th, "With button", func(gtx C) D {
			return lotusui.VStack(th.Space.SM,
				func(gtx C) D { return taSend.LayoutField(th, gtx, "Type your message here.") },
				lotusui.FullWidth(lotusui.Button(th, &taSendBtn, "Send message", lotusui.ButtonProps{})),
			)(gtx)
		}),
	)
}

// ---- toast ----

// One Toaster PER SECTION: each example box owns its stack, so a
// toast fired in one box never appears in another.
var toaster struct {
	ts             [3]lotusui.Toaster
	show, bad      widget.Clickable
	ok, info, warn widget.Clickable
	promise        widget.Clickable
	due            time.Time
}

func toastDemo(th *lotusui.Theme, gtx C) D {
	if !toaster.due.IsZero() && time.Now().After(toaster.due) {
		toaster.due = time.Time{}
		toaster.ts[2].Update("promise", lotusui.Toast{Variant: lotusui.ToastSuccess, Title: "Event created."})
	}
	// Clicks are handled INSIDE each section closure — each box adds
	// to ITS OWN Toaster, so stacks never travel between boxes.
	return card(th, gtx,
		section(th, "Toast — transient, auto-dismissing, stacked", func(gtx C) D {
			if toaster.show.Clicked(gtx) {
				toaster.ts[0].Add(lotusui.Toast{Title: "Event has been created", Description: "Sunday, August 2nd at 4:30pm"})
			}
			if toaster.bad.Clicked(gtx) {
				toaster.ts[0].Add(lotusui.Toast{Variant: lotusui.ToastDestructive, Title: "Save failed", Description: "The server did not respond."})
			}
			return lotusui.HStack(lotusui.Space.SM,
				lotusui.Button(th, &toaster.show, "Show toast", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}),
				lotusui.Button(th, &toaster.bad, "Show error", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}),
			)(gtx)
		}),
		section(th, "Types — success, info, warning", func(gtx C) D {
			if toaster.ok.Clicked(gtx) {
				toaster.ts[1].Add(lotusui.Toast{Variant: lotusui.ToastSuccess, Title: "Event has been created"})
			}
			if toaster.info.Clicked(gtx) {
				toaster.ts[1].Add(lotusui.Toast{Variant: lotusui.ToastInfo, Title: "Arrive 10 minutes before the event"})
			}
			if toaster.warn.Clicked(gtx) {
				toaster.ts[1].Add(lotusui.Toast{Variant: lotusui.ToastWarning, Title: "Event capacity is almost full"})
			}
			return lotusui.HStack(lotusui.Space.SM,
				lotusui.Button(th, &toaster.ok, "Success", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}),
				lotusui.Button(th, &toaster.info, "Info", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}),
				lotusui.Button(th, &toaster.warn, "Warning", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}),
			)(gtx)
		}),
		section(th, "Promise — loading, then the outcome in place", func(gtx C) D {
			if toaster.promise.Clicked(gtx) {
				toaster.ts[2].Add(lotusui.Toast{ID: "promise", Loading: true, Title: "Creating event…", Duration: time.Minute})
				toaster.due = time.Now().Add(2 * time.Second)
			}
			return lotusui.Button(th, &toaster.promise, "Create Event", lotusui.ButtonProps{Variant: lotusui.ButtonOutline})(gtx)
		}),
	)
}

func toastOverlay(th *lotusui.Theme, gtx C) D {
	// Each region renders only ITS section's stack; the full-page demo
	// (no section index) renders them all.
	if i := sectionIndex(demoState); i >= 0 && i < len(toaster.ts) {
		return toaster.ts[i].Layout(th, gtx)
	}
	var d D
	for i := range toaster.ts {
		d = toaster.ts[i].Layout(th, gtx)
	}
	return d
}

// sectionIndex parses the numeric section from a region state; -1
// when the state addresses the whole demo.
func sectionIndex(state string) int {
	n, err := strconv.Atoi(state)
	if err != nil {
		return -1
	}
	return n
}

// ---- toggle & toggle group ----

var (
	tg1, tg2       lotusui.Toggle
	tgSm, tgMd     lotusui.Toggle
	tgLg           lotusui.Toggle
	tgOff1, tgOff2 lotusui.Toggle
	tgg            = lotusui.ToggleGroup{Options: markOpts()}
	tggMulti       = lotusui.ToggleGroup{Multiple: true, Options: []lotusui.ToggleOption{
		{Label: "Bold", Value: "bold", Icon: lotusui.IconTextBold},
		{Label: "Italic", Value: "italic", Icon: lotusui.IconTextItalic},
	}}
)

// markOpts is the bold/italic/underline set — icon-only options MUST
// carry explicit Values, since an empty Label has no value to store.
func markOpts() []lotusui.ToggleOption {
	return []lotusui.ToggleOption{
		{Value: "bold", Icon: lotusui.IconTextBold},
		{Value: "italic", Icon: lotusui.IconTextItalic},
		{Value: "underline", Icon: lotusui.IconTextUnderline},
	}
}

func toggleDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "Toggle — pressed state on the struct", lotusui.HStack(lotusui.Space.SM,
			func(gtx C) D { return tg1.Layout(th, gtx, lotusui.ToggleProps{Icon: lotusui.IconTextBold}) },
			func(gtx C) D {
				return tg2.Layout(th, gtx, lotusui.ToggleProps{Icon: lotusui.IconTextItalic, Label: "Italic", Outline: true})
			},
		)),
		section(th, "ToggleGroup — single select", func(gtx C) D {
			return tgg.Layout(th, gtx, lotusui.SizeMD)
		}),
		section(th, "ToggleGroup — multiple", func(gtx C) D {
			return tggMulti.Layout(th, gtx, lotusui.SizeMD)
		}),
		section(th, "Sizes", lotusui.HStack(lotusui.Space.SM,
			func(gtx C) D {
				return tgSm.Layout(th, gtx, lotusui.ToggleProps{Label: "Small", Outline: true, Size: lotusui.SizeSM})
			},
			func(gtx C) D { return tgMd.Layout(th, gtx, lotusui.ToggleProps{Label: "Default", Outline: true}) },
			func(gtx C) D {
				return tgLg.Layout(th, gtx, lotusui.ToggleProps{Label: "Large", Outline: true, Size: lotusui.SizeLG})
			},
		)),
		section(th, "Disabled", lotusui.HStack(lotusui.Space.SM,
			func(gtx C) D { return tgOff1.Layout(th, gtx, lotusui.ToggleProps{Label: "Disabled", Disabled: true}) },
			func(gtx C) D {
				return tgOff2.Layout(th, gtx, lotusui.ToggleProps{Label: "Disabled", Outline: true, Disabled: true})
			},
		)),
		section(th, "Group outline", func(gtx C) D {
			return tggOutline.Layout(th, gtx, lotusui.SizeMD)
		}),
		section(th, "Group sizes", func(gtx C) D {
			return lotusui.VStack(th.Space.MD,
				func(gtx C) D {
					return tggSizeSM.Layout(th, gtx, lotusui.SizeSM)
				},
				func(gtx C) D {
					return tggSizeMD.Layout(th, gtx, lotusui.SizeMD)
				},
			)(gtx)
		}),
		section(th, "Group disabled", func(gtx C) D {
			return tggOff.Layout(th, gtx, lotusui.SizeMD)
		}),
		section(th, "Group vertical", func(gtx C) D {
			return tggVert.Layout(th, gtx, lotusui.SizeMD)
		}),
		section(th, "Group spacing", func(gtx C) D {
			return tggSpace.Layout(th, gtx, lotusui.SizeSM)
		}),
		section(th, "Font weight selector — custom item content", func(gtx C) D {
			cell := func(sample lotusui.Size, caption string, weight font.Weight) layout.Widget {
				return func(gtx C) D {
					d := gtx.Dp(44)
					gtx.Constraints.Min = image.Pt(d, d)
					return layout.Center.Layout(gtx, func(gtx C) D {
						return lotusui.VStack(2,
							func(gtx C) D {
								l := lotusui.LabelTitle(th, "Aa")
								l.Font.Weight = weight
								return layout.Center.Layout(gtx, l.Layout)
							},
							func(gtx C) D {
								l := lotusui.LabelCaption(th, caption)
								l.Color = th.Palette.FgMuted
								return layout.Center.Layout(gtx, l.Layout)
							},
						)(gtx)
					})
				}
			}
			return lotusui.Field(th, lotusui.FieldProps{Label: "Font Weight", Helper: "Choose the font weight for the interface."}, func(gtx C) D {
				tggWeight.Options = []lotusui.ToggleOption{
					{Value: "light", Content: cell(lotusui.SizeLG, "Light", font.Light)},
					{Value: "normal", Content: cell(lotusui.SizeLG, "Normal", font.Normal)},
					{Value: "medium", Content: cell(lotusui.SizeLG, "Medium", font.Medium)},
					{Value: "bold", Content: cell(lotusui.SizeLG, "Bold", font.Bold)},
				}
				if !tggWeightInit {
					tggWeightInit = true
					tggWeight.SetValue("normal")
				}
				tggWeight.Outline, tggWeight.Spacing = true, 8
				return tggWeight.Layout(th, gtx, lotusui.SizeLG)
			})(gtx)
		}),
	)
}

var (
	sides      = func() []lotusui.ToggleOption { return lotusui.ToggleOpts("Top", "Bottom", "Left", "Right") }
	tggOutline = lotusui.ToggleGroup{Outline: true, Options: lotusui.ToggleOpts("All", "Missed")}
	tggSizeSM  = lotusui.ToggleGroup{Outline: true, Options: sides()}
	tggSizeMD  = lotusui.ToggleGroup{Outline: true, Options: sides()}
	tggOff     = lotusui.ToggleGroup{Multiple: true, Disabled: true, Options: markOpts()}
	tggVert    = lotusui.ToggleGroup{Multiple: true, Vertical: true, Spacing: 1, Options: markOpts()}
	tggSpace   = lotusui.ToggleGroup{Outline: true, Spacing: 8, Options: sides()}
	// The weight selector's options carry Content closures over the
	// theme, so they are rebuilt each frame (the palette can change).
	tggWeight     lotusui.ToggleGroup
	tggWeightInit bool
)

// ---- tooltip ----

var (
	ttip             lotusui.Tooltip
	ttTop, ttRight   lotusui.Tooltip
	ttBottom, ttLeft lotusui.Tooltip
	ttipBtn          widget.Clickable
	ttB1, ttB2       widget.Clickable
	ttB3, ttB4       widget.Clickable
)

func tooltipDemo(th *lotusui.Theme, gtx C) D {
	return card(th, gtx,
		section(th, "Tooltip — rest the pointer on the button", func(gtx C) D {
			return ttip.Layout(th, gtx, "Add to library",
				lotusui.Button(th, &ttipBtn, "Hover me", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}))
		}),
		section(th, "Sides", func(gtx C) D {
			ttTop.Side, ttRight.Side, ttBottom.Side, ttLeft.Side = lotusui.TooltipTop, lotusui.TooltipRight, lotusui.TooltipBottom, lotusui.TooltipLeft
			return lotusui.HStack(lotusui.Space.SM,
				func(gtx C) D {
					return ttTop.Layout(th, gtx, "Add to library", lotusui.Button(th, &ttB1, "Top", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}))
				},
				func(gtx C) D {
					return ttRight.Layout(th, gtx, "Add to library", lotusui.Button(th, &ttB2, "Right", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}))
				},
				func(gtx C) D {
					return ttBottom.Layout(th, gtx, "Add to library", lotusui.Button(th, &ttB3, "Bottom", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}))
				},
				func(gtx C) D {
					return ttLeft.Layout(th, gtx, "Add to library", lotusui.Button(th, &ttB4, "Left", lotusui.ButtonProps{Variant: lotusui.ButtonOutline}))
				},
			)(gtx)
		}),
	)
}

// ---- item ----

var (
	itemAction, itemReview, itemLink1, itemLink2 widget.Clickable
	itemAddA, itemAdd1, itemAdd2, itemAdd3       widget.Clickable
)

func itemDemo(th *lotusui.Theme, gtx C) D {
	icon := func(name string) layout.Widget {
		return lotusui.SVGIcon(name, 18, th.Palette.Fg)
	}
	return card(th, gtx,
		section(th, "Item — title, description, actions", func(gtx C) D {
			return lotusui.VStack(th.Space.MD,
				lotusui.Item(th, lotusui.ItemProps{
					Variant: lotusui.ItemOutline,
					Content: lotusui.ItemContent(th,
						lotusui.ItemTitle(th, "Basic Item"),
						lotusui.ItemDescription(th, "A simple item with title and description."),
					),
					Actions: lotusui.ItemActions(th,
						lotusui.Button(th, &itemAction, "Action", lotusui.ButtonProps{Variant: lotusui.ButtonOutline, Size: lotusui.SizeSM}),
					),
				}),
				lotusui.Item(th, lotusui.ItemProps{
					Variant: lotusui.ItemOutline,
					Size:    lotusui.SizeSM,
					Media:   lotusui.ItemMedia(th, lotusui.ItemMediaDefault, icon(lotusui.IconAccept)),
					Content: lotusui.ItemContent(th, lotusui.ItemTitle(th, "Your profile has been verified.")),
					Actions: lotusui.ItemActions(th, icon(lotusui.IconChevronRight)),
				}),
			)(gtx)
		}),
		section(th, "Variant", func(gtx C) D {
			row := func(v lotusui.ItemVariant, title, desc string) layout.Widget {
				return lotusui.Item(th, lotusui.ItemProps{
					Variant: v,
					Media:   lotusui.ItemMedia(th, lotusui.ItemMediaIcon, icon(lotusui.IconMail)),
					Content: lotusui.ItemContent(th, lotusui.ItemTitle(th, title), lotusui.ItemDescription(th, desc)),
				})
			}
			return lotusui.VStack(th.Space.MD,
				row(lotusui.ItemDefault, "Default Variant", "Transparent background with no border."),
				row(lotusui.ItemOutline, "Outline Variant", "Outlined style with a visible border."),
				row(lotusui.ItemMuted, "Muted Variant", "Muted background for secondary content."),
			)(gtx)
		}),
		section(th, "Size", func(gtx C) D {
			row := func(sz lotusui.Size, title, desc string) layout.Widget {
				return lotusui.Item(th, lotusui.ItemProps{
					Variant: lotusui.ItemOutline,
					Size:    sz,
					Media:   lotusui.ItemMedia(th, lotusui.ItemMediaIcon, icon(lotusui.IconMail)),
					Content: lotusui.ItemContent(th, lotusui.ItemTitle(th, title), lotusui.ItemDescription(th, desc)),
				})
			}
			return lotusui.VStack(th.Space.MD,
				row(lotusui.SizeMD, "Default Size", "The standard size for most use cases."),
				row(lotusui.SizeSM, "Small Size", "A compact size for dense layouts."),
				row(lotusui.SizeXS, "Extra Small Size", "The most compact size available."),
			)(gtx)
		}),
		section(th, "Icon", func(gtx C) D {
			return lotusui.Item(th, lotusui.ItemProps{
				Variant: lotusui.ItemOutline,
				Media:   lotusui.ItemMedia(th, lotusui.ItemMediaIcon, icon(lotusui.IconLock)),
				Content: lotusui.ItemContent(th,
					lotusui.ItemTitle(th, "Security Alert"),
					lotusui.ItemDescription(th, "New login detected from unknown device."),
				),
				Actions: lotusui.ItemActions(th,
					lotusui.Button(th, &itemReview, "Review", lotusui.ButtonProps{Variant: lotusui.ButtonOutline, Size: lotusui.SizeSM}),
				),
			})(gtx)
		}),
		section(th, "Avatar", func(gtx C) D {
			return lotusui.Item(th, lotusui.ItemProps{
				Variant: lotusui.ItemOutline,
				Media: lotusui.ItemMedia(th, lotusui.ItemMediaDefault,
					lotusui.Avatar(th, lotusui.AvatarProps{Initials: "ER"})),
				Content: lotusui.ItemContent(th,
					lotusui.ItemTitle(th, "evilrabbit"),
					lotusui.ItemDescription(th, "evilrabbit@vercel.com"),
				),
				Actions: lotusui.ItemActions(th,
					lotusui.Button(th, &itemAddA, "", lotusui.ButtonProps{Variant: lotusui.ButtonGhost, Size: lotusui.SizeSM, IconStart: lotusui.IconPlus}),
				),
			})(gtx)
		}),
		section(th, "Image", func(gtx C) D {
			img := func(gtx C) D {
				return lotusui.Fill(gtx, th.Palette.BgMuted)
			}
			return lotusui.Item(th, lotusui.ItemProps{
				Variant: lotusui.ItemOutline,
				Media:   lotusui.ItemMedia(th, lotusui.ItemMediaImage, img),
				Content: lotusui.ItemContent(th,
					lotusui.ItemTitle(th, "Album cover"),
					lotusui.ItemDescription(th, "ItemMedia variant=image clips to a rounded square."),
				),
			})(gtx)
		}),
		section(th, "Group", func(gtx C) D {
			people := []struct{ name, email, initials string }{
				{"shadcn", "shadcn@vercel.com", "S"},
				{"maxleiter", "maxleiter@vercel.com", "M"},
				{"evilrabbit", "evilrabbit@vercel.com", "E"},
			}
			btns := []*widget.Clickable{&itemAdd1, &itemAdd2, &itemAdd3}
			var rows []layout.Widget
			for i, p := range people {
				i, p := i, p
				rows = append(rows, lotusui.Item(th, lotusui.ItemProps{
					Variant: lotusui.ItemOutline,
					Media: lotusui.ItemMedia(th, lotusui.ItemMediaDefault,
						lotusui.Avatar(th, lotusui.AvatarProps{Initials: p.initials})),
					Content: lotusui.ItemContent(th,
						lotusui.ItemTitle(th, p.name),
						lotusui.ItemDescription(th, p.email),
					),
					Actions: lotusui.ItemActions(th,
						lotusui.Button(th, btns[i], "", lotusui.ButtonProps{Variant: lotusui.ButtonGhost, Size: lotusui.SizeSM, IconStart: lotusui.IconPlus}),
					),
				}))
			}
			return lotusui.ItemGroup(th, rows...)(gtx)
		}),
		section(th, "Header", func(gtx C) D {
			tile := func(title, desc string) layout.Widget {
				return lotusui.Item(th, lotusui.ItemProps{
					Variant: lotusui.ItemOutline,
					Header: lotusui.ItemHeader(th, func(gtx C) D {
						gtx.Constraints.Min.Y = gtx.Dp(72)
						return lotusui.Fill(gtx, th.Palette.BgMuted)
					}, nil),
					Content: lotusui.ItemContent(th,
						lotusui.ItemTitle(th, title),
						lotusui.ItemDescription(th, desc),
					),
				})
			}
			return lotusui.HStack(th.Space.SM,
				tile("v0-1.5-sm", "Everyday tasks and UI generation."),
				tile("v0-1.5-lg", "Advanced thinking or reasoning."),
				tile("v0-2.0-mini", "Open Source model for everyone."),
			)(gtx)
		}),
		section(th, "Link", func(gtx C) D {
			return lotusui.VStack(th.Space.SM,
				lotusui.Item(th, lotusui.ItemProps{
					Btn: &itemLink1,
					Content: lotusui.ItemContent(th,
						lotusui.ItemTitle(th, "Visit our documentation"),
						lotusui.ItemDescription(th, "Learn how to get started with our components."),
					),
					Actions: lotusui.ItemActions(th, icon(lotusui.IconChevronRight)),
				}),
				lotusui.Item(th, lotusui.ItemProps{
					Variant: lotusui.ItemOutline,
					Btn:     &itemLink2,
					Content: lotusui.ItemContent(th,
						lotusui.ItemTitle(th, "External resource"),
						lotusui.ItemDescription(th, "Opens in a new tab with security attributes."),
					),
					Actions: lotusui.ItemActions(th, icon(lotusui.IconShare)),
				}),
			)(gtx)
		}),
	)
}
