# Installation

`mani` is available on Linux and Mac, with partial support for Windows.

* Binaries are available on the [release](https://github.com/alajmo/mani/releases) page

* via cURL (Linux & macOS)
  ```bash
  curl -sfL https://raw.githubusercontent.com/alajmo/mani/main/install.sh | sh
  ```

* via Homebrew
  ```bash
  brew tap alajmo/mani
  brew install mani
  ```

* via Arch
  ```sh
  pacman -S mani
  ```

* via Nix
  ```sh
  nix-env -iA nixos.mani
  ```

* via Go
  ```bash
  go get -u github.com/alajmo/mani
  ```

## Building From Source

1. Clone the repo
2. Build and run the executable

    ```bash
    make build && ./dist/mani
    ```
