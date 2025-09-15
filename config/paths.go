package config

import (
	"os"
	"path/filepath"
)

type AssetPaths struct {
	BaseDir   string
	Images    string
	Gifs      string
	Videos    string
	Documents string
	Audio     string
	Stickers  string
}

func NewAssetPaths() *AssetPaths {
	baseDir := os.Getenv("ASSETS_PATH")
	if baseDir == "" {
		baseDir = "./assets"
	}

	return &AssetPaths{
		BaseDir: baseDir,
		//Images:    filepath.Join(baseDir, "images"),
		Gifs: filepath.Join(baseDir, "gifs"),
		//Videos:    filepath.Join(baseDir, "videos"),
		//Documents: filepath.Join(baseDir, "documents"),
		//Audio:     filepath.Join(baseDir, "audio"),
		//Stickers:  filepath.Join(baseDir, "stickers"),
	}
}
