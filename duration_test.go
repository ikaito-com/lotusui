package lotusui

import (
	"image/color"
	"testing"
	"time"
)

func TestDurationDefaults(t *testing.T) {
	th := NewTheme()
	if th.Duration.Fast != 150*time.Millisecond {
		t.Fatalf("Fast: got %v", th.Duration.Fast)
	}
	if th.Duration.Normal != 200*time.Millisecond {
		t.Fatalf("Normal: got %v", th.Duration.Normal)
	}
	if th.Duration.Slow != 300*time.Millisecond {
		t.Fatalf("Slow: got %v", th.Duration.Slow)
	}
}

func TestWithDuration(t *testing.T) {
	prev := Duration
	t.Cleanup(func() { Duration = prev })
	th := NewTheme(WithDuration(DurationScale{
		Fast:   50 * time.Millisecond,
		Normal: 80 * time.Millisecond,
		Slow:   120 * time.Millisecond,
	}))
	if th.Duration.Fast != 50*time.Millisecond {
		t.Fatalf("theme Fast: %v", th.Duration.Fast)
	}
	if Duration.Fast != 50*time.Millisecond {
		t.Fatalf("package Duration not synced: %v", Duration.Fast)
	}
}

func TestLerpNRGBA(t *testing.T) {
	a := color.NRGBA{}
	b := color.NRGBA{R: 100, G: 200, B: 50, A: 255}
	mid := lerpNRGBA(a, b, 0.5)
	if mid.R != 50 || mid.G != 100 || mid.B != 25 {
		t.Fatalf("mid RGB = %+v", mid)
	}
	if mid.A != 127 && mid.A != 128 {
		t.Fatalf("mid A = %d", mid.A)
	}
	if lerpNRGBA(a, b, 0) != a || lerpNRGBA(a, b, 1) != b {
		t.Fatal("endpoints")
	}
}
