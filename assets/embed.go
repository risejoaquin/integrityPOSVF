package assets

import "embed"

//go:embed default/templates default/static
var DefaultAssets embed.FS
