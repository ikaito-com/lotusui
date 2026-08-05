DIST := dist

.PHONY: build docsapp docsapp-data gallery-wasm run-docsapp serve serve-docsapp \
	serve-gallery clean check goldens media

# Default deploy artifact: docsapp WASM at site root (GitHub Pages).
build: docsapp ## Build docsapp for Pages into $(DIST)/

docsapp-data: ## Sync versions/bench/changelog into docsapp/data for WASM embed
	@mkdir -p docsapp/data
	cp -f versions.json docsapp/data/versions.json
	cp -f bench.json docsapp/data/bench.json
	cp -f ../CHANGELOG.md docsapp/data/CHANGELOG.md
	@echo "  docsapp/data synced (heroes stay in media/ — not embedded)"

docsapp: docsapp-data ## Build the lotusui docs app WASM into $(DIST)/ (Pages root)
	@# Wipe stale HTML-site leftovers (components/, style.css, …) so
	# local serve matches a clean GitHub Pages deploy.
	rm -rf $(DIST)
	@mkdir -p $(DIST)/media
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $(DIST)/docsapp.wasm ./docsapp
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(DIST)/wasm_exec.js
	cp docsapp/index.html $(DIST)/index.html
	cp favicon.svg $(DIST)/favicon.svg
	cp versions.json $(DIST)/versions.json
	cp -f media/hero-light.png media/hero-dark.png media/devices.png media/colors.png $(DIST)/media/
	@echo "  clean Pages artifact in $(DIST)/  (make serve)"
	@ls -la $(DIST)/ | head -20

gallery-wasm: ## Gallery WASM only (goldens / media screenshots)
	@mkdir -p $(DIST)/gallery
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $(DIST)/gallery/gallery.wasm ./gallery
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(DIST)/gallery/wasm_exec.js
	cp gallery/index.html $(DIST)/gallery/index.html
	cp favicon.svg $(DIST)/favicon.svg

# PAGE=select opens that docs route; default is home.
# Do not default to a heavy component page — button alone has ~15
# live Previews and used to freeze the first frame at startup.
# Build to bin/ then exec so "stuck" is never just a quiet go run compile.
PAGE ?=
BINDIR := bin
run-docsapp: ## Run the docsapp natively (optional PAGE=button|select|…)
	@mkdir -p $(BINDIR)
	@echo "building docsapp…"
	go build -o $(BINDIR)/docsapp ./docsapp
	@if [ -n "$(PAGE)" ]; then \
		echo "docsapp → #$(PAGE)"; \
		LOTUSUI_DOCS=$(PAGE) ./$(BINDIR)/docsapp; \
	else \
		echo "docsapp → home"; \
		LOTUSUI_DOCS= ./$(BINDIR)/docsapp; \
	fi

# GitHub Pages artifact: thin index.html + docsapp.wasm + media/ under dist/.
serve: build ## Build Pages artifact and serve on :3030 (same as CI deploy)
	@echo "GitHub Pages preview → http://localhost:3030/"
	go run ./cmd/serve -dir $(DIST) -addr :3030

serve-docsapp: serve ## Alias of serve

serve-gallery: gallery-wasm ## Serve gallery WASM on :3031 (screenshots)
	go run ./cmd/serve -dir $(DIST) -addr :3031

clean:
	rm -rf $(DIST) $(BINDIR)

CHROME ?= /Applications/Google Chrome.app/Contents/MacOS/Google Chrome
STATES := palette scales typography stack button button/loading input \
          checkbox switch select tabs dialog dialog/open list badge dropdown-menu card grid simplegrid \
          split split/depth1 icons

goldens: gallery-wasm ## Headless gallery screenshots → goldens/
	@mkdir -p goldens
	@(go run ./cmd/serve -dir $(DIST) -addr :3111 >/dev/null 2>&1 &) ; sleep 2
	@for s in $(STATES); do \
		f=goldens/$$(echo $$s | tr / -).png; \
		echo "  #$$s -> $$f"; \
		"$(CHROME)" --headless=new --window-size=900,700 --virtual-time-budget=20000 \
			--screenshot=$$f "http://localhost:3111/gallery/#$$s" 2>/dev/null; \
	done
	@pkill -f "cmd/serve -dir $(DIST) -addr :3111" || true

media: gallery-wasm ## README / homepage screenshots from gallery showcase
	@mkdir -p media
	@(go run ./cmd/serve -dir $(DIST) -addr :3112 >/dev/null 2>&1 &) ; sleep 2
	@"$(CHROME)" --headless=new --window-size=850,480 --force-device-scale-factor=2 --virtual-time-budget=20000 \
		--screenshot=media/hero-light.png "http://localhost:3112/gallery/#showcase" 2>/dev/null
	@"$(CHROME)" --headless=new --window-size=850,480 --force-device-scale-factor=2 --virtual-time-budget=20000 \
		--screenshot=media/hero-dark.png "http://localhost:3112/gallery/#showcase&palette=midnight" 2>/dev/null
	@"$(CHROME)" --headless=new --window-size=850,360 --force-device-scale-factor=2 --virtual-time-budget=20000 \
		--screenshot=media/colors.png "http://localhost:3112/gallery/#showcase-colors" 2>/dev/null
	@"$(CHROME)" --headless=new --window-size=880,470 --force-device-scale-factor=2 --virtual-time-budget=20000 \
		--screenshot=media/devices.png "http://localhost:3112/gallery/#showcase-devices" 2>/dev/null
	@pkill -f "cmd/serve -dir $(DIST) -addr :3112" || true
	@ls -la media/

check: ## Site module CI: fmt, vet, test, build native + wasm
	gofmt -l .
	go vet ./...
	go test ./docspages/ ./docsapp/
	go build -o /dev/null ./gallery ./live ./docsapp ./docspages ./cmd/serve
	GOOS=js GOARCH=wasm go build -o /dev/null ./gallery
	GOOS=js GOARCH=wasm go build -o /dev/null ./docsapp
