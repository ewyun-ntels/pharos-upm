package ui

//go:generate pnpm install
//go:generate pnpm run build

import (
	"embed"
	"io/fs"
)

//go:embed all:out
var distDir embed.FS

// DistDirFS contains the embedded out directory files
var DistDirFS, _ = fs.Sub(distDir, "out")

//go:embed all:images
var UiConfigImagesFS embed.FS
