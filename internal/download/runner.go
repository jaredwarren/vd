package download

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Runner executes yt-dlp on the host or via Docker.
type Runner struct {
	DownloadDir string
	DockerImage string
}

// Run downloads the video at rawURL and returns the host filesystem path of the output file.
func (r *Runner) Run(ctx context.Context, rawURL string, jobStarted time.Time) (string, error) {
	ytdlp, err := exec.LookPath("yt-dlp")
	if err == nil {
		return r.runHost(ctx, ytdlp, rawURL, jobStarted)
	}
	return r.runDocker(ctx, rawURL, jobStarted)
}

func commonArgs(outputPattern string) []string {
	return []string{
		"-f", "bv*+ba/b",
		"--merge-output-format", "mp4",
		"-o", outputPattern,
		"--no-progress",
		"--newline",
		"--print", "after_move:filepath",
	}
}

func (r *Runner) runHost(ctx context.Context, ytdlp, rawURL string, jobStarted time.Time) (string, error) {
	pattern := filepath.Join(r.DownloadDir, "%(title)s.%(ext)s")
	args := append(commonArgs(pattern), rawURL)
	cmd := exec.CommandContext(ctx, ytdlp, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", formatErr(err, stderr.String())
	}
	if p := strings.TrimSpace(stdout.String()); p != "" {
		if line := lastNonEmptyLine(p); line != "" {
			return line, nil
		}
	}
	return newestFileSince(r.DownloadDir, jobStarted, stderr.String())
}

func (r *Runner) runDocker(ctx context.Context, rawURL string, jobStarted time.Time) (string, error) {
	img := r.DockerImage
	if img == "" {
		img = "jauderho/yt-dlp"
	}
	pattern := "/downloads/%(title)s.%(ext)s"
	args := []string{
		"run", "--rm",
		"-v", r.DownloadDir + ":/downloads",
		img,
	}
	args = append(args, commonArgs(pattern)...)
	args = append(args, rawURL)
	cmd := exec.CommandContext(ctx, "docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", formatErr(err, stderr.String())
	}
	out := strings.TrimSpace(stdout.String())
	if out != "" {
		if line := lastNonEmptyLine(out); line != "" {
			return dockerPathToHost(line, r.DownloadDir), nil
		}
	}
	return newestFileSince(r.DownloadDir, jobStarted, stderr.String())
}

func dockerPathToHost(containerPath, hostDir string) string {
	p := strings.TrimSpace(containerPath)
	const prefix = "/downloads/"
	if strings.HasPrefix(p, prefix) {
		return filepath.Join(hostDir, filepath.FromSlash(strings.TrimPrefix(p, prefix)))
	}
	return p
}

func formatErr(err error, stderr string) error {
	msg := strings.TrimSpace(stderr)
	if len(msg) > 400 {
		msg = msg[len(msg)-400:]
	}
	if msg == "" {
		return err
	}
	return fmt.Errorf("%w: %s", err, msg)
}

func lastNonEmptyLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if t := strings.TrimSpace(lines[i]); t != "" {
			return t
		}
	}
	return ""
}

func newestFileSince(dir string, notBefore time.Time, stderrHint string) (string, error) {
	var best string
	var bestTime time.Time
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().Before(notBefore) {
			return nil
		}
		if info.ModTime().After(bestTime) {
			bestTime = info.ModTime()
			best = path
		}
		return nil
	})
	if best != "" {
		return best, nil
	}
	msg := strings.TrimSpace(stderrHint)
	if len(msg) > 300 {
		msg = msg[len(msg)-300:]
	}
	if msg != "" {
		return "", fmt.Errorf("download finished but output path unknown: %s", msg)
	}
	return "", fmt.Errorf("download finished but output path unknown")
}
