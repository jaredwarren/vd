package web

import "embed"

// Assets are served by the HTTP server (PWA shell).
//
//go:embed index.html manifest.webmanifest sw.js icons/icon-192.png icons/icon-512.png
var Assets embed.FS
