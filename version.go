package lotusui

// The registry manifest is generated from the source at BUILD time
// (and drift-checked by `lotusui verify -registry`); the CLI and AI
// agents consume it — app code never reads it at runtime.
//go:generate go run ./cmd/lotusui registry

// Version is the lotusui release this build comes from — the same
// version the changelog and the docs site are organized by.
const Version = "0.3.4"
