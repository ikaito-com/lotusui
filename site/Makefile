DIST := dist

.PHONY: build gen wasm serve clean check

build: gen wasm ## Full site build into $(DIST)

gen: ## Generate the static HTML pages
	go run ./gen -out $(DIST)

wasm: ## Compile the gallery to WebAssembly (one bundle for the whole site)
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $(DIST)/gallery/gallery.wasm ./gallery
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(DIST)/gallery/wasm_exec.js

serve: build ## Build, then serve locally on :3030
	go run ./gen -out $(DIST) -serve :3030

clean:
	rm -rf $(DIST)

# Golden screenshots: every gallery state, captured headlessly. Run
# `make goldens` after a visual change and review the diff like code —
# a component's look becomes part of its reviewed contract. (CI can
# upload these as artifacts; exact-pixel gating across machines is
# deliberately not attempted — font rendering differs per platform.)
CHROME ?= /Applications/Google Chrome.app/Contents/MacOS/Google Chrome
STATES := palette scales typography stack button button/loading input \
          checkbox switch select tabs dialog dialog/open list badge dropdown-menu card grid simplegrid \
          split split/depth1 icons

goldens: build
	@mkdir -p goldens
	@(go run ./gen -out $(DIST) -serve :3111 >/dev/null 2>&1 &) ; sleep 2
	@for s in $(STATES); do \
		f=goldens/$$(echo $$s | tr / -).png; \
		echo "  #$$s -> $$f"; \
		"$(CHROME)" --headless=new --window-size=900,700 --virtual-time-budget=20000 \
			--screenshot=$$f "http://localhost:3111/gallery/#$$s" 2>/dev/null; \
	done
	@pkill -f "gen -out $(DIST) -serve :3111" || true

# Showcase media: the README / homepage screenshots, captured from the
# REAL wasm gallery (states in gallery/showcase.go). Re-run after any
# visual change so every marketing image stays truthful.
media: build
	@mkdir -p media
	@(go run ./gen -out $(DIST) -serve :3112 >/dev/null 2>&1 &) ; sleep 2
	@"$(CHROME)" --headless=new --window-size=850,480 --force-device-scale-factor=2 --virtual-time-budget=20000 \
		--screenshot=media/hero-light.png "http://localhost:3112/gallery/#showcase" 2>/dev/null
	@"$(CHROME)" --headless=new --window-size=850,480 --force-device-scale-factor=2 --virtual-time-budget=20000 \
		--screenshot=media/hero-dark.png "http://localhost:3112/gallery/#showcase&palette=midnight" 2>/dev/null
	@"$(CHROME)" --headless=new --window-size=850,360 --force-device-scale-factor=2 --virtual-time-budget=20000 \
		--screenshot=media/colors.png "http://localhost:3112/gallery/#showcase-colors" 2>/dev/null
	@"$(CHROME)" --headless=new --window-size=880,470 --force-device-scale-factor=2 --virtual-time-budget=20000 \
		--screenshot=media/devices.png "http://localhost:3112/gallery/#showcase-devices" 2>/dev/null
	@pkill -f "gen -out $(DIST) -serve :3112" || true
	@ls -la media/

check: ## What CI runs for the site module: everything compiles, both targets
	gofmt -l .
	go vet ./...
	go build -o /dev/null ./gallery
	GOOS=js GOARCH=wasm go build -o /dev/null ./gallery
