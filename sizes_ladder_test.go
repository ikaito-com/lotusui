package lotusui

import "testing"

// Every size ladder must be STRICTLY monotonic from 2XS to 2XL — a
// flattened step reads as "this size is not defined". This test is
// the tripwire that keeps every preset visually distinct.
func TestSizeLaddersStrictlyMonotonic(t *testing.T) {
	order := []Size{Size2XS, SizeXS, SizeSM, SizeMD, SizeLG, SizeXL, Size2XL}

	assertInc := func(name string, vals []float32) {
		t.Helper()
		for i := 1; i < len(vals); i++ {
			if vals[i] <= vals[i-1] {
				t.Errorf("%s ladder not strictly increasing at step %d: %v", name, i, vals)
				return
			}
		}
	}

	var btnH, inH, cbox, swW, bdgH, avat, cardP []float32
	for _, sz := range order {
		ratio, v, _ := ButtonProps{Size: sz}.metrics()
		btnH = append(btnH, ratio*16*1.3+2*float32(v)) // text + padding ≈ height
		r2, v2, _ := inputMetrics(sz)
		inH = append(inH, r2*16*1.3+2*float32(v2))
		cbox = append(cbox, float32((&Checkbox{Size: sz}).boxDp()))
		w, _ := (&Switch{Size: sz}).trackDp()
		swW = append(swW, float32(w))
		avat = append(avat, float32(avatarDp(sz)))
		cardP = append(cardP, float32(CardProps{Size: sz}.Pad()))
		// Badge: ratio+vpad proxy via the same switch Badge uses.
		switch sz {
		case Size2XS:
			bdgH = append(bdgH, 9+2*1)
		case SizeXS:
			bdgH = append(bdgH, 10+2*2)
		case SizeSM:
			bdgH = append(bdgH, 11+2*2.5)
		case SizeMD:
			bdgH = append(bdgH, 12+2*3)
		case SizeLG:
			bdgH = append(bdgH, 13+2*4)
		case SizeXL:
			bdgH = append(bdgH, 14+2*5)
		case Size2XL:
			bdgH = append(bdgH, 15+2*6)
		}
	}
	assertInc("Button", btnH)
	assertInc("Input", inH)
	assertInc("Checkbox", cbox)
	assertInc("Switch", swW)
	assertInc("Avatar", avat)
	assertInc("Card", cardP)
	assertInc("Badge", bdgH)
}
