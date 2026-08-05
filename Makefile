ICONS_DIR := assets/icons

.PHONY: icons release generate verify test vet fmt tidy check

icons: ## Download the SVG icons per $(ICONS_DIR)/manifest.txt from the Iconify API (see ICONS.md; commit the results — builds never need the network)
	go run ./cmd/lotusui icons -manifest $(ICONS_DIR)/manifest.txt -out $(ICONS_DIR) -gen icons_gen.go -genpkg lotusui

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
