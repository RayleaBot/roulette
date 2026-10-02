package roulette

import _ "embed"

// Manifest contains the same defaults used by the host and management page.
//
//go:embed info.json
var Manifest []byte
