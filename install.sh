#!/bin/sh
#
# initrule installer (macOS / Linux).
#
# curl -fsSL https://raw.githubusercontent.com/mewisme/initrule/main/install.sh | sh
#
# Uninstall: curl -fsSL .../install.sh | sh -s -- --uninstall
#
# Environment:
#   INITRULE_VERSION     release tag (default: latest)
#   INITRULE_INSTALL_DIR bundle location (default: ~/.initrule)
#   INITRULE_BIN_DIR     symlink location (default: ~/.local/bin)
set -eu

REPO="mewisme/initrule"
INSTALL_DIR="${INITRULE_INSTALL_DIR:-$HOME/.initrule}"
BIN_DIR="${INITRULE_BIN_DIR:-$HOME/.local/bin}"

if [ "${1:-}" = "--uninstall" ]; then
	rm -f "$BIN_DIR/initrule"
	rm -rf "$INSTALL_DIR"
	echo "initrule uninstalled (removed $INSTALL_DIR and $BIN_DIR/initrule)."
	exit 0
fi

os="$(uname -s)"
arch="$(uname -m)"
case "$os" in
	Darwin) os="darwin" ;;
	Linux) os="linux" ;;
	*) echo "initrule: unsupported OS '$os'." >&2; exit 1 ;;
esac
case "$arch" in
	arm64|aarch64) arch="arm64" ;;
	x86_64|amd64) arch="amd64" ;;
	i386|i686|x86) arch="386" ;;
	# GoReleaser archive name for GOARCH=arm GOARM=7
	armv7l|armv7) arch="armv7" ;;
	*) echo "initrule: unsupported architecture '$arch'." >&2; exit 1 ;;
esac
if [ "$os" = "darwin" ] && { [ "$arch" = "386" ] || [ "$arch" = "armv7" ]; }; then
	echo "initrule: unsupported architecture '$arch' on darwin." >&2
	exit 1
fi

version="${INITRULE_VERSION:-}"
if [ -z "$version" ]; then
	version="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest" \
		| sed -n 's#.*/releases/tag/##p')"
fi
if [ -z "$version" ]; then
	version="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
		| sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)"
fi
[ -n "$version" ] || {
	echo "initrule: could not resolve latest version; set INITRULE_VERSION (e.g. INITRULE_VERSION=v0.0.4)." >&2
	exit 1
}
case "$version" in v*) ;; *) version="v$version" ;; esac
ver="${version#v}"

url="https://github.com/$REPO/releases/download/$version/initrule_${ver}_${os}_${arch}.tar.gz"
echo "Installing initrule $version ($os/$arch)..."
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
curl -fsSL "$url" -o "$tmp/initrule.tar.gz" || {
	echo "initrule: download failed: $url" >&2
	exit 1
}

dest="$INSTALL_DIR/versions/$version"
rm -rf "$dest"
mkdir -p "$dest"
tar -xzf "$tmp/initrule.tar.gz" -C "$dest"
chmod +x "$dest/initrule"

mkdir -p "$BIN_DIR"
ln -sf "$dest/initrule" "$BIN_DIR/initrule"
ln -sfn "$dest" "$INSTALL_DIR/current"

# Prune older versions.
if [ -d "$INSTALL_DIR/versions" ]; then
	for d in "$INSTALL_DIR/versions"/*; do
		[ -d "$d" ] || continue
		[ "$d" = "$dest" ] || rm -rf "$d"
	done
fi

echo "Installed to $dest"
echo "Linked $BIN_DIR/initrule"

on_path=0
oldifs="$IFS"
IFS=:
for dir in $PATH; do
	[ "$dir" = "$BIN_DIR" ] && on_path=1 && break
done
IFS="$oldifs"

if [ "$on_path" -eq 0 ]; then
	echo ""
	echo "$BIN_DIR is not on your PATH. Add it:"
	echo "  export PATH=\"$BIN_DIR:\$PATH\""
fi

echo ""
echo "Done. Run: initrule --help"
