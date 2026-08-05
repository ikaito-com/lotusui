//go:build darwin

package lotusui

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>
#include <stdint.h>

static void applySeamless(NSView *view, int attempts);

// showTrafficLights re-shows the three standard buttons. Gio >= v0.8
// implements app.Decorated(false) natively on macOS (transparent
// titlebar, hidden title, full-size content — what applyProps does)
// but it also HIDES the traffic lights; keeping them is the one part
// of the seamless look Gio doesn't offer. Idempotent.
static void showTrafficLights(NSWindow *w) {
	[w standardWindowButton:NSWindowCloseButton].hidden = NO;
	[w standardWindowButton:NSWindowMiniaturizeButton].hidden = NO;
	[w standardWindowButton:NSWindowZoomButton].hidden = NO;
}

// applyProps is the flash killer: the titlebar-hiding properties are
// safe to set synchronously at view-attach time (before the first
// present), so the decorated bar never draws even for one frame.
// Gio's own Configure (from app.Decorated(false)) sets the same
// properties, but possibly after the first present — this keeps the
// no-flash guarantee pinned here.
static void applyProps(NSWindow *w) {
	w.titlebarAppearsTransparent = YES;
	w.titleVisibility = NSWindowTitleHidden;
	w.styleMask |= NSWindowStyleMaskFullSizeContentView;
	showTrafficLights(w);
}

// reassertButtons wins the race with Gio's Configure: the
// app.Decorated(false) option processing hides the standard buttons
// and can land after us, so re-assert a few times over the first
// couple of seconds. hidden=NO is idempotent — no visual side effect
// once settled.
static void reassertButtons(NSView *view, int attempts) {
	NSWindow *w = view.window;
	if (w != nil) {
		showTrafficLights(w);
	}
	if (attempts > 0) {
		dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.2 * NSEC_PER_SEC)),
			dispatch_get_main_queue(), ^{ reassertButtons(view, attempts - 1); });
	}
}

static void applySeamless(NSView *view, int attempts) {
	NSWindow *w = view.window;
	if (w == nil) {
		// The view can reach us before it's attached to its window —
		// retry briefly instead of silently doing nothing.
		if (attempts > 0) {
			dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.15 * NSEC_PER_SEC)),
				dispatch_get_main_queue(), ^{ applySeamless(view, attempts - 1); });
		}
		return;
	}
	{
		// The title bar disappears as a bar but its traffic lights stay,
		// floating over our own chrome, and content extends to the very
		// top — no "window" feeling. The (now transparent) titlebar strip
		// still drags the window.
		applyProps(w);
		// An empty unified toolbar is the sanctioned way to get the
		// inset, comfortably-padded traffic lights (instead of them
		// hugging the very corner): it heightens the invisible titlebar
		// region and AppKit repositions the buttons for it — stable
		// across resizes, no frame hacks.
		NSToolbar *tb = [[NSToolbar alloc] initWithIdentifier:@"lotusui-chrome"];
		tb.showsBaselineSeparator = NO;
		w.toolbar = tb;
		if (@available(macOS 11.0, *)) {
			w.toolbarStyle = NSWindowToolbarStyleUnified;
		}
		// Force a REAL relayout: setting the same frame is a no-op in
		// AppKit, so nudge the height by a point and put it back — the
		// classic trick that makes the titlebar strip vanish immediately
		// instead of lingering until the first manual resize.
		NSRect f = w.frame;
		NSRect nudged = f;
		nudged.size.height += 1;
		[w setFrame:nudged display:YES];
		[w setFrame:f display:YES];
		reassertButtons(view, 10);
	}
}

static void seamless(uintptr_t viewRef) {
	NSView *view = (__bridge NSView *)(void *)viewRef;
	// Gio delivers AppKitViewEvent on the main thread, BEFORE the first
	// frame is presented — applying synchronously here is what prevents
	// the decorated titlebar from flashing for a split second at launch.
	// The async path is only the fallback for a not-yet-attached view.
	if ([NSThread isMainThread] && view.window != nil) {
		// Synchronously hide the titlebar BEFORE the first present (no
		// flash) — but never mutate frames/toolbars from inside Gio's
		// own init event, which tears the window down. Those follow on
		// the next runloop tick.
		applyProps(view.window);
	}
	dispatch_async(dispatch_get_main_queue(), ^{
		applySeamless(view, 10);
	});
}
*/
import "C"

// MakeSeamlessWindow hides the macOS title bar (keeping the traffic
// lights) and lets content occupy the full window — called from main's
// event loop on every AppKitViewEvent. The window MUST also be created
// with app.Decorated(false): on Gio >= v0.8 that both provides the
// native titlebar-hiding and — critically — stops Gio drawing its own
// fallback decorations (it reads NSWindowStyleMaskFullSizeContentView
// back from the window and would otherwise paint a CSD title bar over
// ours). This function then adds what Decorated(false) doesn't: the
// visible traffic lights and the unified-toolbar inset.
func MakeSeamlessWindow(view uintptr) {
	if view == 0 {
		return
	}
	C.seamless(C.uintptr_t(view))
}
