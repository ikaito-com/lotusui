module github.com/ikaito-com/lotusui

go 1.26.1

// v0.4.0–v0.8.0 belong to an ABANDONED lineage: this history was
// squashed and re-tagged from v0.1.0, but the module proxy caches the
// pre-squash versions forever. They are a different codebase (no
// item.go, no ScrollArea), so resolving them reads as a downgrade with
// missing API. Always upgrade to an explicit tag.
//
// NOTE: cmd/go only honours retractions from the module's HIGHEST
// version, so while the ghost v0.8.0 is the highest this directive is
// inert and `go get @latest` still resolves to it. It becomes
// effective the day a tag above v0.8.0 is published.
retract [v0.4.0, v0.8.0]

require (
	gioui.org v0.10.1
	github.com/srwiley/oksvg v0.0.0-20221011165216-be6e8873101c
	github.com/srwiley/rasterx v0.0.0-20220730225603-2ab79fcdd4ef
)

require (
	github.com/go-text/typesetting v0.3.4 // indirect
	golang.org/x/exp/shiny v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/image v0.26.0 // indirect
	golang.org/x/net v0.54.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
)
