#!/bin/sh
# Credit to https://github.com/ducaale/xh for this install script.

set -e

if [ "$(uname -s)" = "Darwin" ] && [ "$(uname -m)" = "x86_64" ]; then
    target="darwin_amd64"
elif [ "$(uname -s)" = "Linux" ] && [ "$(uname -m)" = "x86_64" ]; then
    target="linux_amd64"
elif [ "$(uname -s)" = "Linux" ] && ( uname -m | grep -q -e '^arm' -e '^aarch' ); then
    target="linux_arm64"
else
    echo "Unsupported OS or architecture"
    exit 1
fi

fetch() {
    if which curl > /dev/null; then
        if [ "$#" -eq 2 ]; then curl -L -o "$1" "$2"; else curl -sSL "$1"; fi
    elif which wget > /dev/null; then
        if [ "$#" -eq 2 ]; then wget -O "$1" "$2"; else wget -nv -O - "$1"; fi
    else
        echo "Can't find curl or wget, can't download package"
        exit 1
    fi
}

echo "Detected target: $target"

url=$(
    fetch https://api.github.com/repos/alajmo/mani/releases/latest |
    tac | tac | grep -wo -m1 "https://.*$target.tar.gz" || true
)

if ! test "$url"; then
    echo "Could not find release info"
    exit 1
fi

echo "Downloading mani..."

temp_dir=$(mktemp -dt mani.XXXXXX)
trap 'rm -rf "$temp_dir"' EXIT INT TERM
cd "$temp_dir"

if ! fetch mani.tar.gz "$url"; then
    echo "Could not download tarball"
    exit 1
fi

user_bin="$HOME/.local/bin"
case $PATH in
    *:"$user_bin":* | "$user_bin":* | *:"$user_bin")
        default_bin=$user_bin
        ;;
    *)
        default_bin='/usr/local/bin'
        ;;
esac

printf "Install location [default: %s]: " "$default_bin"
read -r bindir < /dev/tty
bindir=${bindir:-$default_bin}

while ! test -d "$bindir"; do
    echo "Directory $bindir does not exist"
    printf "Install location [default: %s]: " "$default_bin"
    read -r bindir < /dev/tty
    bindir=${bindir:-$default_bin}
done

tar xzf mani.tar.gz

if test -w "$bindir"; then
    mv mani "$bindir/"
else
    sudo mv mani "$bindir/"
fi

$bindir/mani version
