package main

import (
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/jaredwarren/ytdl/internal/config"
	"github.com/jaredwarren/ytdl/internal/download"
	"github.com/jaredwarren/ytdl/internal/jobs"
	"github.com/jaredwarren/ytdl/internal/server"
	siteweb "github.com/jaredwarren/ytdl/web"
)

func main() {
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	runner := &download.Runner{
		DownloadDir: cfg.DownloadDir,
		DockerImage: cfg.DockerImage,
	}
	store := jobs.NewStore(runner)

	mux := http.NewServeMux()
	server.NewAPI(store).Register(mux)

	fileSrv := http.FileServer(http.FS(siteweb.Assets))
	mux.Handle("/", server.LogMiddleware(fileSrv))

	addr := cfg.Listen
	log.Printf("listening on %s, downloads to %s", addr, cfg.DownloadDir)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
