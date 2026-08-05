//go:build !js

package main

import (
	"image"
	_ "image/png"
	"os"
	"sync"

	"gioui.org/op/paint"
	"golang.org/x/image/draw"
)

var (
	heroOnce sync.Once
	heroSrc  []image.Image

	heroScaledMu sync.Mutex
	heroScaled   = map[int][]paint.ImageOp{} // display width → ops
)

func loadHeroImages() []paint.ImageOp {
	// Width 0 = source-sized ops (legacy callers); prefer heroOpsFor(width).
	return heroOpsFor(0)
}

// heroOpsFor returns hero ImageOps sized for display width px.
// Scaling is done on the CPU once per width — never op.Affine under
// ScrollArea (Affine + clip + scroll offset collapses the macOS window
// to 0×0 and spins the GPU at 100% CPU).
func heroOpsFor(width int) []paint.ImageOp {
	heroOnce.Do(func() {
		names := []string{"hero-light.png", "hero-dark.png", "devices.png"}
		for _, n := range names {
			if img := decodeHeroFile(n); img != nil {
				heroSrc = append(heroSrc, img)
			}
		}
	})
	if len(heroSrc) == 0 {
		return nil
	}
	if width < 1 {
		width = 900
	}
	heroScaledMu.Lock()
	defer heroScaledMu.Unlock()
	if ops, ok := heroScaled[width]; ok {
		return ops
	}
	ops := make([]paint.ImageOp, 0, len(heroSrc))
	for _, img := range heroSrc {
		ops = append(ops, paint.NewImageOp(scaleHeroTo(img, width)))
	}
	heroScaled[width] = ops
	return ops
}

// startHeroFetch is a no-op on native — heroes load from media/.
func startHeroFetch(func()) {}

func decodeHeroFile(name string) image.Image {
	for _, dir := range []string{"media/", "../media/"} {
		f, err := os.Open(dir + name)
		if err != nil {
			continue
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err == nil {
			return img
		}
	}
	return nil
}

func scaleHeroTo(img image.Image, maxW int) image.Image {
	b := img.Bounds()
	if b.Dx() <= maxW {
		return img
	}
	nh := b.Dy() * maxW / b.Dx()
	dst := image.NewRGBA(image.Rect(0, 0, maxW, nh))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
	return dst
}
