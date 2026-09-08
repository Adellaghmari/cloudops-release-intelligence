package policies

import "embed"

//go:embed VERSION release_gate.rego
var Bundle embed.FS
