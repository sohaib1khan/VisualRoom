#!/usr/bin/env bash
# Install VisualRoom (Vroom) on Linux. Works without a local Go toolchain.
# Author: Sohaib Khan  https://github.com/sohaib1khan/VisualRoom
#
#   curl -fsSL https://raw.githubusercontent.com/sohaib1khan/VisualRoom/main/install.sh | bash
#
# Or from a clone:
#
#   ./install.sh
#
# Override the install location with PREFIX (default: ~/.local/bin).

set -euo pipefail

REPO="${VROOM_REPO:-sohaib1khan/VisualRoom}"
APP="vroom"
GO_VERSION="${VROOM_GO_VERSION:-1.22.8}"
PREFIX="${PREFIX:-}"
KEEP_GO="${VROOM_KEEP_GO:-0}"
INSTALL_WORK=""

log() { printf 'vroom-install: %s\n' "$*"; }
die() { printf 'vroom-install: ERROR: %s\n' "$*" >&2; exit 1; }

cleanup() {
	if [[ -n "${INSTALL_WORK}" && "${KEEP_GO}" != "1" ]]; then
		rm -rf "${INSTALL_WORK}"
	fi
}
trap cleanup EXIT

need_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "need '$1' on PATH"
}

detect_arch() {
	local m
	m="$(uname -m)"
	case "$m" in
		x86_64 | amd64) echo amd64 ;;
		aarch64 | arm64) echo arm64 ;;
		*) die "unsupported architecture: $m (need amd64 or arm64)" ;;
	esac
}

os_name() {
	local s
	s="$(uname -s)"
	case "$s" in
		Linux) echo linux ;;
		*) die "this installer targets Linux (found $s)" ;;
	esac
}

go_ok() {
	command -v go >/dev/null 2>&1 || return 1
	local major minor
	# GOVERSION is like go1.22.8
	local ver
	ver="$(go env GOVERSION 2>/dev/null || go version | awk '{print $3}')"
	ver="${ver#go}"
	major="${ver%%.*}"
	minor="${ver#*.}"
	minor="${minor%%.*}"
	[[ "$major" =~ ^[0-9]+$ && "$minor" =~ ^[0-9]+$ ]] || return 1
	if ((major > 1)); then
		return 0
	fi
	((minor >= 22))
}

pick_prefix() {
	if [[ -n "$PREFIX" ]]; then
		echo "$PREFIX"
		return
	fi
	if [[ -d /usr/local/bin && -w /usr/local/bin ]]; then
		echo /usr/local/bin
		return
	fi
	echo "${HOME}/.local/bin"
}

in_source_tree() {
	[[ -f go.mod && -f main.go ]] && grep -q '^module vroom$' go.mod
}

download() {
	local url="$1" dest="$2"
	if command -v curl >/dev/null 2>&1; then
		curl -fL --retry 3 --retry-delay 1 -o "$dest" "$url"
	elif command -v wget >/dev/null 2>&1; then
		wget -O "$dest" "$url"
	else
		die "need curl or wget to download $url"
	fi
}

fetch_release_binary() {
	local os="$1" arch="$2" dest="$3"
	local api asset url
	api="https://api.github.com/repos/${REPO}/releases/latest"
	local json=""
	if command -v curl >/dev/null 2>&1; then
		json="$(curl -fsSL "$api" 2>/dev/null || true)"
	elif command -v wget >/dev/null 2>&1; then
		json="$(wget -qO- "$api" 2>/dev/null || true)"
	fi
	[[ -n "$json" ]] || return 1
	asset="${APP}-${os}-${arch}"
	url="$(printf '%s' "$json" | grep '"browser_download_url"' | grep "/${asset}\"" | head -n1 | cut -d '"' -f 4)"
	[[ -n "$url" ]] || return 1
	log "downloading ${asset} from GitHub releases"
	download "$url" "$dest"
	chmod +x "$dest"
}

bootstrap_go() {
	local os="$1" arch="$2" dir="$3"
	local tarball="go${GO_VERSION}.${os}-${arch}.tar.gz"
	local url="https://dl.google.com/go/${tarball}"
	log "bootstrapping Go ${GO_VERSION} (${os}/${arch})"
	mkdir -p "$dir"
	download "$url" "${dir}/${tarball}"
	tar -C "$dir" -xzf "${dir}/${tarball}"
	export GOROOT="${dir}/go"
	export GOTOOLCHAIN=local
	export PATH="${GOROOT}/bin:${PATH}"
	go version >/dev/null
}

fetch_source() {
	local dir="$1"
	mkdir -p "$dir"
	if command -v git >/dev/null 2>&1; then
		log "cloning https://github.com/${REPO}.git"
		git clone --depth 1 "https://github.com/${REPO}.git" "$dir/src"
	else
		log "downloading source tarball"
		download "https://github.com/${REPO}/archive/refs/heads/main.tar.gz" "$dir/src.tgz"
		mkdir -p "$dir/extract"
		tar -C "$dir/extract" -xzf "$dir/src.tgz"
		mv "$dir"/extract/* "$dir/src"
	fi
}

build_from_source() {
	local src="$1" dest="$2"
	log "building static binary (CGO_ENABLED=0)"
	(
		cd "$src"
		CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$dest" .
	)
	chmod +x "$dest"
}

main() {
	need_cmd uname
	need_cmd tar
	need_cmd mkdir
	need_cmd grep
	need_cmd chmod

	local os arch prefix bin src
	os="$(os_name)"
	arch="$(detect_arch)"
	prefix="$(pick_prefix)"
	mkdir -p "$prefix"
	bin="${prefix}/${APP}"

	INSTALL_WORK="$(mktemp -d "${TMPDIR:-/tmp}/vroom-install.XXXXXX")"

	if fetch_release_binary "$os" "$arch" "${INSTALL_WORK}/${APP}" 2>/dev/null; then
		install -m 755 "${INSTALL_WORK}/${APP}" "$bin"
	else
		log "no prebuilt release for ${os}/${arch}; building from source"
		if ! go_ok; then
			if command -v go >/dev/null 2>&1; then
				log "system Go is too old ($(go env GOVERSION 2>/dev/null || go version)); need 1.22+"
			else
				log "no Go toolchain on PATH"
			fi
			bootstrap_go "$os" "$arch" "$INSTALL_WORK/go-sdk"
		else
			log "using $(go env GOVERSION) from PATH"
		fi

		if in_source_tree; then
			src="$(pwd)"
			log "building from $(pwd)"
		else
			fetch_source "$INSTALL_WORK"
			src="$INSTALL_WORK/src"
		fi
		build_from_source "$src" "${INSTALL_WORK}/${APP}"
		if command -v install >/dev/null 2>&1; then
			install -m 755 "${INSTALL_WORK}/${APP}" "$bin"
		else
			cp "${INSTALL_WORK}/${APP}" "$bin"
			chmod 755 "$bin"
		fi
	fi

	log "installed ${bin}"
	if ! command -v "$APP" >/dev/null 2>&1; then
		log "add to your PATH, then re-open the shell:"
		log "  export PATH=\"${prefix}:\$PATH\""
		case "${SHELL:-}" in
			*/zsh) log "  echo 'export PATH=\"${prefix}:\$PATH\"' >> ~/.zshrc" ;;
			*/fish) log "  fish_add_path ${prefix}" ;;
			*) log "  echo 'export PATH=\"${prefix}:\$PATH\"' >> ~/.bashrc" ;;
		esac
	fi
	log "try:  ${APP} preview"
}

main "$@"
