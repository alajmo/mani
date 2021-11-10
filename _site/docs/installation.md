# Installation

`mani` is available on Linux and Mac, with partial support for Windows.

* Binaries are available on the [release](https://github.com/alajmo/mani/releases) page

* via cURL (Linux & macOS)
  ```sh
  curl -sfL https://raw.githubusercontent.com/alajmo/mani/main/install.sh | sh
  ```

* via Homebrew
  ```sh
  brew tap alajmo/mani
  brew install mani
  ```

* via Go
  ```sh
  go get -u github.com/alajmo/mani
  ```

## Building From Source

1. Clone the repo
2. Build and run the executable
    ```sh
    make build && ./dist/mani
    ```
