# ytdl

Small **Go** HTTP service with a **mobile-friendly PWA**: paste a **YouTube** URL, and the server runs **yt-dlp** in the background and saves the file under a directory you configure. Intended for **LAN** use on a home server (e.g. **Debian** + **systemd**).

## Features

- Web UI + installable PWA (manifest + service worker)
- Background jobs with status polling
- **Host `yt-dlp`** if it is on `PATH`, otherwise **`docker run`** with [`jauderho/yt-dlp`](https://hub.docker.com/r/jauderho/yt-dlp)
- Output filenames use the video title (`%(title)s.%(ext)s`), merged to **mp4** when formats split
- Server-side checks: only YouTube hosts are accepted

## Requirements

- **Go 1.22+** (to build)
- For actual downloads, either:
  - **[yt-dlp](https://github.com/yt-dlp/yt-dlp)** installed and on `PATH`, or
  - **Docker** (e.g. Docker Desktop on macOS) for the fallback image

## Quick start (local)

```bash
cd ytdl   # repository root
make run
```

Open [http://127.0.0.1:8080](http://127.0.0.1:8080). By default, files go to `./tmp/downloads`. Override:

```bash
make run YTD_LISTEN=:8080 YTD_DOWNLOAD_DIR=~/Downloads/ytdl
```

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `YTD_DOWNLOAD_DIR` | **yes** | — | Absolute or relative path; created if missing |
| `YTD_LISTEN` | no | `:8080` | Listen address (e.g. `127.0.0.1:8080` or `:8080`) |
| `YTD_DOCKER_IMAGE` | no | `jauderho/yt-dlp` | Image used when `yt-dlp` is not on `PATH` |

Example:

```bash
export YTD_DOWNLOAD_DIR=/var/lib/ytdl/downloads
export YTD_LISTEN=:8080
./ytdl
```

## Makefile

| Target | Description |
|--------|-------------|
| `make` / `make build` | Build `bin/ytdl` for the current OS/arch |
| `make run` | Run via `go run` with defaults (see above) |
| `make linux-amd64` | Cross-build `bin/ytdl-linux-amd64` for Linux **x86_64** (Intel/AMD 64-bit, e.g. Xeon) |
| `make vet` `fmt` `test` `tidy` | Standard Go maintenance |
| `make clean` | Remove `bin/` |
| `make help` | Short usage |

On Debian, `uname -m` should show `x86_64` for that Linux binary. (Use `GOARCH=arm64` only for 64-bit ARM machines.)

## HTTP API

- `POST /api/jobs` — body: `{"url":"https://www.youtube.com/..."}` → `201` with `{ "id", "status" }`
- `GET /api/jobs/{id}` — `{ "status", "message?", "output_path?", ... }`  
  Status values: `queued`, `running`, `completed`, `failed`

Jobs are stored **in memory**; they are lost when the process restarts.

## Debian + systemd

1. Build or copy the Linux binary, e.g. to `/usr/local/bin/ytdl`.
2. Create a user and download directory (see comments in [`deploy/ytdl.service`](deploy/ytdl.service)).
3. Install **yt-dlp** for that user, **or** add the user to the `docker` group so `docker run` works non-interactively.
4. Install the unit:

   ```bash
   sudo cp deploy/ytdl.service /etc/systemd/system/ytdl.service
   sudo systemctl daemon-reload
   sudo systemctl enable --now ytdl
   ```

Edit `Environment=` lines in the unit file as needed.

## Security

This is meant for a **trusted LAN**: there is **no TLS or login** in v1. You should still only paste **YouTube** URLs you trust; the server rejects non-YouTube hosts. For exposure beyond the LAN, use a reverse proxy, TLS, and authentication separately.

## License

No license file is included; add one if you publish the repo.
