package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds runtime settings from the environment.
type Config struct {
	Listen      string
	DownloadDir string
	DockerImage string
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	c := Config{
		Listen:      getenv("YTD_LISTEN", ":8080"),
		DownloadDir: strings.TrimSpace(os.Getenv("YTD_DOWNLOAD_DIR")),
		DockerImage: getenv("YTD_DOCKER_IMAGE", "jauderho/yt-dlp"),
	}
	if c.DownloadDir == "" {
		return Config{}, fmt.Errorf("YTD_DOWNLOAD_DIR is required")
	}
	abs, err := filepath.Abs(c.DownloadDir)
	if err != nil {
		return Config{}, fmt.Errorf("YTD_DOWNLOAD_DIR: %w", err)
	}
	c.DownloadDir = abs
	if st, err := os.Stat(c.DownloadDir); err != nil {
		if os.IsNotExist(err) {
			if mkErr := os.MkdirAll(c.DownloadDir, 0o755); mkErr != nil {
				return Config{}, fmt.Errorf("create download dir: %w", mkErr)
			}
		} else {
			return Config{}, fmt.Errorf("YTD_DOWNLOAD_DIR: %w", err)
		}
	} else if !st.IsDir() {
		return Config{}, fmt.Errorf("YTD_DOWNLOAD_DIR is not a directory")
	}
	return c, nil
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
