package lotusui

import (
	"image"
	"time"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// ToastVariant selects the toast's role.
type ToastVariant int

const (
	ToastDefault     ToastVariant = iota
	ToastDestructive              // danger ink for failures
	ToastSuccess                  // success ink
	ToastInfo                     // info ink
	ToastWarning                  // warning ink
)

// Toast is one transient notification.
type Toast struct {
	Title       string
	Description string
	Variant     ToastVariant
	// ID names the toast so Update can replace it in place — the
	// promise pattern: Add a Loading toast, Update it on completion.
	ID string
	// Loading renders a spinner before the title.
	Loading bool
	// Duration before auto-dismiss; zero means 4s.
	Duration time.Duration
}

type activeToast struct {
	Toast
	start time.Time
}

// Toaster owns the toast queue: Add from anywhere, Layout ONCE per
// frame at WINDOW constraints (the same portal rule as Dialog) — the
// stack renders bottom-right, newest at the bottom, each auto-
// dismissing after its duration. Self-invalidates while non-empty.
type Toaster struct {
	items []activeToast
}

// Add enqueues a toast. Safe to call during event handling.
func (t *Toaster) Add(toast Toast) {
	if toast.Duration == 0 {
		toast.Duration = 4 * time.Second
	}
	t.items = append(t.items, activeToast{Toast: toast})
}

// Update replaces the live toast with the given ID in place and
// restarts its clock — the promise pattern's completion. A miss adds
// the toast instead.
func (t *Toaster) Update(id string, toast Toast) {
	if toast.Duration == 0 {
		toast.Duration = 4 * time.Second
	}
	toast.ID = id
	for i := range t.items {
		if t.items[i].ID == id {
			t.items[i] = activeToast{Toast: toast}
			return
		}
	}
	t.items = append(t.items, activeToast{Toast: toast})
}

func (t *Toaster) Layout(th *Theme, gtx layout.Context) layout.Dimensions {
	if len(t.items) == 0 {
		return layout.Dimensions{}
	}
	// Expire from the front; stamp start times on first sight (Add has
	// no clock — gtx.Now is the only time source, idle-safe).
	live := t.items[:0]
	for _, it := range t.items {
		if it.start.IsZero() {
			it.start = gtx.Now
		}
		if gtx.Now.Sub(it.start) < it.Duration {
			live = append(live, it)
		}
	}
	t.items = live
	if len(t.items) == 0 {
		return layout.Dimensions{}
	}
	// One wake-up at the NEXT expiry rather than a frame every vsync:
	// nothing here animates between deadlines (a Loading toast's
	// spinner drives its own frames).
	next := t.items[0].start.Add(t.items[0].Duration)
	for _, it := range t.items[1:] {
		if d := it.start.Add(it.Duration); d.Before(next) {
			next = d
		}
	}
	gtx.Execute(op.InvalidateCmd{At: next})

	margin := gtx.Dp(th.Space.LG)
	toastW := gtx.Dp(340)
	if max := gtx.Constraints.Max.X - 2*margin; toastW > max {
		toastW = max
	}
	y := gtx.Constraints.Max.Y - margin
	for i := len(t.items) - 1; i >= 0; i-- {
		it := t.items[i]
		m := op.Record(gtx.Ops)
		cgtx := gtx
		cgtx.Constraints = layout.Constraints{Min: image.Pt(toastW, 0), Max: image.Pt(toastW, 1<<20)}
		dims := layout.UniformInset(th.Space.MD).Layout(cgtx, func(gtx layout.Context) layout.Dimensions {
			ink, body := th.Palette.Fg, th.Palette.FgMuted
			switch it.Variant {
			case ToastDestructive:
				ink, body = th.Palette.Danger, th.Palette.Danger
			case ToastSuccess:
				ink = th.Palette.Success
			case ToastInfo:
				ink = th.Palette.Info
			case ToastWarning:
				ink = th.Palette.Warning
			}
			title := LabelBody(th, it.Title)
			title.Color = ink
			title.Font.Weight = font.Medium
			head := title.Layout
			if it.Loading {
				head = func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(Spinner(th, 14)),
						layout.Rigid(HSpacer(th.Space.SM)),
						layout.Rigid(title.Layout),
					)
				}
			}
			if it.Description == "" {
				return head(gtx)
			}
			desc := LabelCaption(th, it.Description)
			desc.Color = body
			return VStack(4, head, desc.Layout)(gtx)
		})
		call := m.Stop()
		size := image.Pt(toastW, dims.Size.Y)
		y -= size.Y
		st := op.Offset(image.Pt(gtx.Constraints.Max.X-margin-toastW, y)).Push(gtx.Ops)
		r := gtx.Dp(th.Radius.MD)
		cardShadow(gtx, size, r)
		cl := clip.UniformRRect(image.Rectangle{Max: size}, r).Push(gtx.Ops)
		paint.Fill(gtx.Ops, th.Palette.BgPanel)
		widget.Border{Color: th.Palette.Border, Width: unit.Dp(1), CornerRadius: th.Radius.MD}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: size} })
		call.Add(gtx.Ops)
		cl.Pop()
		st.Pop()
		y -= gtx.Dp(th.Space.SM)
	}
	return layout.Dimensions{Size: gtx.Constraints.Max}
}
