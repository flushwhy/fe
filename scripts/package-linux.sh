#!/usr/bin/env bash
set -euo pipefail

version="${1:?Usage: package-linux.sh VERSION [CHANNEL] [ARCH]}"
channel="${2:-}"
target_arch="${3:-all}"
case "$target_arch" in
  all|amd64|arm64) ;;
  *) echo "Unsupported architecture: $target_arch" >&2; exit 2 ;;
esac
mkdir -p dist/packages
deb_version="$version"
rpm_release="1"
archive_suffix="$version"
if [[ -n "$channel" ]]; then
  deb_version="${version}~${channel}"
  rpm_release="${channel//[^A-Za-z0-9._]/.}"
  archive_suffix="${version}-${channel}"
fi

targets="amd64 arm64"
if [[ "$target_arch" != all ]]; then targets="$target_arch"; fi
for target in $targets; do
  binary="dist/fe-linux-$target"
  test -f "$binary"
  chmod 0755 "$binary"
  case "$target" in
    amd64) deb_arch="amd64"; rpm_arch="x86_64" ;;
    arm64) deb_arch="arm64"; rpm_arch="aarch64" ;;
  esac

  package="dist/packages/deb-$target"
  mkdir -p "$package/DEBIAN" "$package/usr/bin"
  install -m 0755 "$binary" "$package/usr/bin/fe"
  cat > "$package/DEBIAN/control" <<EOF
Package: fe
Version: $deb_version
Section: devel
Priority: optional
Architecture: $deb_arch
Maintainer: flushwhy
Description: Local middleware CLI for game and software development
 Glue between your engine/compiler and shipping workflows.
EOF
  dpkg-deb --root-owner-group --build "$package" "dist/packages/fe_${archive_suffix}_$target.deb"
  rm -rf "$package"

  top="$(mktemp -d)"
  mkdir -p "$top"/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}
  cp "$binary" "$top/SOURCES/fe"
  cat > "$top/SPECS/fe.spec" <<EOF
Name: fe
Version: $version
Release: $rpm_release
Summary: Local middleware CLI for game and software development
License: MIT

%description
Glue between your engine/compiler and shipping workflows.

%install
mkdir -p %{buildroot}/usr/bin
install -m 0755 %{_sourcedir}/fe %{buildroot}/usr/bin/fe

%files
/usr/bin/fe
EOF
  # BuildArch in the spec overrides cross-target selection; select it only via --target.
  # The host strip tool cannot process a foreign-architecture Go binary.
  rpm_strip="/usr/bin/strip"
  if [[ "$rpm_arch" == "aarch64" ]]; then rpm_strip="/bin/true"; fi
  rpmbuild --target "$rpm_arch" -bb --define "_topdir $top" --define "__strip $rpm_strip" "$top/SPECS/fe.spec" >/dev/null
  rpm_file="$(find "$top/RPMS" -type f -name '*.rpm' -print -quit)"
  test -n "$rpm_file"
  test "$(rpm -qp --qf '%{ARCH}' "$rpm_file")" = "$rpm_arch"
  cp "$rpm_file" "dist/packages/fe-${archive_suffix}-$rpm_arch.rpm"
  rm -rf "$top"

  archive="$(mktemp -d)"
  mkdir -p "$archive/fe"
  install -m 0755 "$binary" "$archive/fe/fe"
  cat > "$archive/fe/README.txt" <<EOF
Fe command-line tool
Install by copying the fe executable to a directory on your PATH, such as /usr/local/bin.
EOF
  tar -C "$archive" -czf "dist/packages/fe_${archive_suffix}_linux_$target.tar.gz" fe
  rm -rf "$archive"
done
