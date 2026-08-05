ICONS_DIR := assets/icons

.PHONY: icons release generate verify test vet fmt tidy check bench-doc run-docsapp

icons: ## Download the SVG icons per $(ICONS_DIR)/manifest.txt from the Iconify API (see ICONS.md; commit the results — builds never need the network)
	go run ./cmd/lotusui icons -manifest $(ICONS_DIR)/manifest.txt -out $(ICONS_DIR) -gen icons_gen.go -genpkg lotusui

bench-doc: ## Refresh site/bench.json (Performance page numbers) from local go test -bench; optional WASM size if site/dist/gallery/gallery.wasm exists
	@wasm=""; \
	if [ -f site/dist/gallery/gallery.wasm ]; then wasm="-wasm site/dist/gallery/gallery.wasm"; \
	elif [ -f /tmp/lotusui-gallery.wasm ]; then wasm="-wasm /tmp/lotusui-gallery.wasm"; fi; \
	go run ./cmd/lotusui bench-doc -o site/bench.json $$wasm

# Native Gio docs window (site/docsapp). Home by default.
# Examples: make run-docsapp   make run-docsapp PAGE=select
run-docsapp: ## Run the lotusui docsapp natively
	$(MAKE) -C site run-docsapp $(if $(PAGE),PAGE=$(PAGE),)

release: ## Cut a release: make release [BUMP=minor|patch|major] — validates, edits changelog/version/docs manifest, prints the git finish
	go run ./cmd/lotusui release $(if $(BUMP),-bump $(BUMP),)

generate: ## Run all build-time codegen (icons, scales) via the standard go:generate directives
	go generate ./...

verify: ## Offline drift check: generated code must match its sources (fails CI when someone forgot go generate)
	go run ./cmd/lotusui verify -manifest $(ICONS_DIR)/manifest.txt -out $(ICONS_DIR) -gen icons_gen.go -api api.txt -registry

test: ## Run all tests
	go test ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format all Go source
	gofmt -l -w .

tidy: ## Tidy go.mod/go.sum
	go mod tidy

check: fmt vet verify test ## Everything that must pass before committing
