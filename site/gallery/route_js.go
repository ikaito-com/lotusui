//go:build js

package main

import (
	"image"
	"syscall/js"
)

// currentRoute reads the page's URL hash — the docs pages address a
// component/state by iframe src ("gallery/#modal/open").
func currentRoute() string {
	return js.Global().Get("location").Get("hash").String()
}

// onRouteChange re-renders on hashchange so a docs page (or a person
// editing the URL) can switch states without reloading the wasm.
func onRouteChange(invalidate func()) {
	js.Global().Get("window").Call("addEventListener", "hashchange",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			invalidate()
			return nil
		}))
}

// initPalette applies the docs site's saved palette (same-origin
// localStorage) before the first frame, then listens for live switches
// posted by the picker in the site's top bar.
func initPalette(apply func(slug string)) {
	if v := js.Global().Get("localStorage").Call("getItem", "lotusui-palette"); v.Truthy() {
		apply(v.String())
	}
	js.Global().Get("window").Call("addEventListener", "message",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) == 0 {
				return nil
			}
			if p := args[0].Get("data").Get("lotusuiPalette"); p.Truthy() {
				apply(p.String())
			}
			return nil
		}))
}

// initLook applies the saved look at boot and listens for live
// switches from the picker — the second theming axis beside palettes.
func initLook(apply func(slug string)) {
	if v := js.Global().Get("localStorage").Call("getItem", "lotusui-look"); v.Truthy() {
		apply(v.String())
	}
	js.Global().Get("window").Call("addEventListener", "message",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) == 0 {
				return nil
			}
			if l := args[0].Get("data").Get("lotusuiLook"); l.Truthy() {
				apply(l.String())
			}
			return nil
		}))
}

// setRegionScroll updates the wheel router's view of one region:
// enc bit0 = internally scrollable, bit1 = at start, bit2 = at end.
func setRegionScroll(i int, enc int8) {
	js.Global().Call("lotusuiSetRegionScroll", i, int(enc))
}

// initWheelRouter implements scroll chaining OUT of the canvas: the
// browser can't do it for us (the canvas preventDefaults wheels), so
// a capture-phase listener decides per wheel — inner scrollable that
// isn't at its edge keeps the event; everything else forwards the
// scroll to the docs page. Preview iframes have no overlay regions;
// when embedded (parent !== window) every non-inner wheel is chained
// to parent. Standalone /gallery/#… (top-level, no regions) keeps
// native scrolling.
func initWheelRouter() {
	js.Global().Call("eval", `
(function(){
  var regions = [];
  window.lotusuiSetRegions = function(rs){ regions = rs.map(function(r){ return {x:r.x,y:r.y,w:r.w,h:r.h,enc:0}; }); };
  window.lotusuiSetRegionScroll = function(i, enc){ if (regions[i]) regions[i].enc = enc; };
  function deltaY(ev){
    if (ev.deltaMode === 1) return ev.deltaY * 16;
    if (ev.deltaMode === 2) return ev.deltaY * (window.innerHeight || 800);
    return ev.deltaY;
  }
  function chainToPage(ev){
    var scroller = (parent !== window) ? parent : window;
    scroller.scrollBy(0, deltaY(ev));
    ev.preventDefault();
    ev.stopPropagation();
  }
  window.addEventListener("wheel", function(ev){
    var embedded = parent !== window;
    if (!regions.length) {
      // Docs Preview iframe: canvas ate the wheel — scroll the article.
      if (embedded) chainToPage(ev);
      return;
    }
    // Regions are canvas-local (0,0 = demobox). Map client → canvas.
    var canvas = document.querySelector("#giowindow canvas") || document.querySelector("canvas");
    if (!canvas) { if (embedded) chainToPage(ev); return; }
    var br = canvas.getBoundingClientRect();
    var lx = ev.clientX - br.left, ly = ev.clientY - br.top;
    var hit = null;
    for (var i = 0; i < regions.length; i++) {
      var r = regions[i];
      if (lx >= r.x && lx <= r.x + r.w && ly >= r.y && ly <= r.y + r.h) { hit = r; break; }
    }
    var scrollable = hit && (hit.enc & 1);
    var atStart = hit && (hit.enc & 2);
    var atEnd = hit && (hit.enc & 4);
    var inner = scrollable && ((ev.deltaY > 0 && !atEnd) || (ev.deltaY < 0 && !atStart));
    if (!inner) chainToPage(ev);
  }, {capture: true, passive: false});
})();`)
}

// initOverlay wires the docs page's overlay protocol: measure
// requests and region layouts arrive by postMessage (CSS px,
// converted to device px here); a ready signal tells the page the
// wasm is live.
func initOverlay(invalidate func()) {
	dpr := func() float64 {
		d := js.Global().Get("devicePixelRatio").Float()
		if d <= 0 {
			return 1
		}
		return d
	}
	js.Global().Get("window").Call("addEventListener", "message",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) == 0 {
				return nil
			}
			data := args[0].Get("data")
			if m := data.Get("lotusuiMeasure"); m.Truthy() {
				n := m.Length()
				specs := make([]string, n)
				widths := make([]int, n)
				for i := 0; i < n; i++ {
					it := m.Index(i)
					specs[i] = it.Get("state").String()
					widths[i] = int(it.Get("w").Float()*dpr() + 0.5)
				}
				pendingMeasureSpecs, pendingMeasureWidths = specs, widths
				invalidate()
			}
			if rg := data.Get("lotusuiRegions"); rg.Truthy() {
				if f := js.Global().Get("lotusuiSetRegions"); f.Truthy() {
					f.Invoke(rg)
				}
				n := rg.Length()
				regions := make([]overlayRegion, n)
				for i := 0; i < n; i++ {
					it := rg.Index(i)
					slug, state, _, _ := parseRoute(it.Get("state").String())
					x := int(it.Get("x").Float()*dpr() + 0.5)
					y := int(it.Get("y").Float()*dpr() + 0.5)
					w := int(it.Get("w").Float()*dpr() + 0.5)
					h := int(it.Get("h").Float()*dpr() + 0.5)
					regions[i] = overlayRegion{slug: slug, state: state,
						rect: image.Rect(x, y, x+w, y+h)}
				}
				setOverlayRegions(regions)
				invalidate()
			}
			if vs := data.Get("lotusuiVisible"); vs.Truthy() {
				n := vs.Length()
				vis := make(map[int]bool, n)
				for i := 0; i < n; i++ {
					vis[vs.Index(i).Int()] = true
				}
				visibleRegions = vis
				invalidate()
			}
			return nil
		}))
	// Prefer window: docs host WASM in-page (parent === window). The
	// standalone /gallery/ shell and legacy iframe embeds still work —
	// parent is the same window or the embedding page.
	postToDocs(map[string]any{"lotusuiReady": true})
}

func postToDocs(msg map[string]any) {
	js.Global().Get("window").Call("postMessage", msg, "*")
}

// sendMeasured replies to a measure request with natural heights in
// CSS px.
func sendMeasured(px []int) {
	dpr := js.Global().Get("devicePixelRatio").Float()
	if dpr <= 0 {
		dpr = 1
	}
	arr := make([]any, len(px))
	for i, p := range px {
		arr[i] = int(float64(p)/dpr + 0.5)
	}
	postToDocs(map[string]any{"lotusuiHeights": arr})
}

// notifyPainted tells the docs page a frame was drawn for the current
// overlay region. The page snapshots THAT demobox only (never copies
// the frame onto other Previews) and reveals the host once the canvas
// matches the box it mounted into.
func notifyPainted() {
	msg := map[string]any{"lotusuiPainted": true}
	if len(overlayRegions) > 0 {
		r := overlayRegions[0]
		msg["lotusuiRegion"] = r.slug + "/" + r.state
	}
	postToDocs(msg)
}

var lastReportedHeight int

// reportHeight posts the demo's content height (in CSS pixels) to the
// embedding docs page, which resizes the iframe to match — the demo
// gets exactly its height and the page keeps the only scrollbar.
func reportHeight(px int) {
	if px == lastReportedHeight || px <= 0 {
		return
	}
	lastReportedHeight = px
	dpr := js.Global().Get("devicePixelRatio").Float()
	if dpr <= 0 {
		dpr = 1
	}
	css := int(float64(px)/dpr + 0.5)
	postToDocs(map[string]any{"lotusuiHeight": css})
}
