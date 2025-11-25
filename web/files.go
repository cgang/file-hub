package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var webFiles embed.FS

// StaticFiles returns an http.FileSystem for the embedded static files.
func StaticFiles() (http.FileSystem, error) {
	assets, err := fs.Sub(webFiles, "dist")
	if err != nil {
		return nil, err
	}

	return http.FS(assets), nil
}
