#!/usr/bin/env bash
# Install ytdl on Debian (or similar) per deploy/ytdl.service.
#
# Usage (from repo root, as root):
#   sudo ./deploy/install.sh                 # build with local `go` and install
#   sudo ./deploy/install.sh /path/to/ytdl   # install a pre-built binary (e.g. from make linux-amd64)
#
# Optional environment:
#   INSTALL_ADD_DOCKER_GROUP=1   add user `ytdl` to group `docker` (for Docker fallback instead of host yt-dlp)

set -euo pipefail

TARGET_BIN="/usr/local/bin/ytdl"
SERVICE_NAME="ytdl.service"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}"
DATA_ROOT="/var/lib/ytdl"
DOWNLOAD_DIR="${DATA_ROOT}/downloads"

die() {
	echo "error: $*" >&2
	exit 1
}

if [[ "${EUID:-0}" -ne 0 ]]; then
	die "run as root, e.g. sudo $0"
fi

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
UNIT_SRC="${SCRIPT_DIR}/ytdl.service"

[[ -f "${UNIT_SRC}" ]] || die "missing unit file: ${UNIT_SRC}"

ensure_user_and_dirs() {
	if ! getent passwd ytdl >/dev/null; then
		useradd --system \
			--home-dir "${DATA_ROOT}" \
			--shell /usr/sbin/nologin \
			--user-group \
			ytdl
	fi
	mkdir -p "${DOWNLOAD_DIR}"
	chown -R ytdl:ytdl "${DATA_ROOT}"
}

install_binary() {
	local src="${1:-}"
	if [[ -n "${src}" ]]; then
		[[ -f "${src}" ]] || die "binary not found: ${src}"
		install -m 0755 -o root -g root "${src}" "${TARGET_BIN}"
		return
	fi
	if command -v go >/dev/null 2>&1; then
		echo "building ${TARGET_BIN} from ${REPO_ROOT} ..."
		( cd "${REPO_ROOT}" && CGO_ENABLED=0 go build -o "${TARGET_BIN}" ./cmd/ytdl )
		chown root:root "${TARGET_BIN}"
		chmod 0755 "${TARGET_BIN}"
		return
	fi
	die "no binary path given and 'go' not in PATH. Pass the binary: sudo $0 /path/to/ytdl"
}

install_unit() {
	install -m 0644 "${UNIT_SRC}" "${SERVICE_PATH}"
	systemctl daemon-reload
	systemctl enable "${SERVICE_NAME}"
	systemctl restart "${SERVICE_NAME}"
}

optional_docker_group() {
	if [[ "${INSTALL_ADD_DOCKER_GROUP:-0}" == "1" ]]; then
		if getent group docker >/dev/null; then
			usermod -aG docker ytdl
			echo "Added user 'ytdl' to group 'docker'. Reboot (or restart the service after a fresh login) if docker membership was just applied."
		else
			echo "warning: group 'docker' not found; install Docker Engine or skip INSTALL_ADD_DOCKER_GROUP."
		fi
	fi
}

usage() {
	cat <<EOF
Usage: sudo $0 [ /path/to/ytdl-binary ]

With no argument, builds from the repository using Go (must be installed on this machine).
With a path, copies that binary to ${TARGET_BIN}.

Environment:
  INSTALL_ADD_DOCKER_GROUP=1   usermod -aG docker ytdl

After install, either install yt-dlp where the service user can run it, or use Docker + docker group.
EOF
}

case "${1:-}" in
-h | --help)
	usage
	exit 0
	;;
esac

ensure_user_and_dirs
install_binary "${1:-}"
optional_docker_group
install_unit

echo
echo "Installed ${TARGET_BIN} and ${SERVICE_PATH}."
echo "Status: systemctl status ${SERVICE_NAME}"
echo "Logs:   journalctl -u ${SERVICE_NAME} -f"
echo
echo "Downloads go to ${DOWNLOAD_DIR}. Open http://<this-host>:8080 on your LAN (see YTD_LISTEN in the unit file)."
