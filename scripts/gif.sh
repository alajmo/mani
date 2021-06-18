#!/bin/bash

set -eum

MANI_PATH=$(dirname $(dirname $(realpath "$0")))
MANI_EXAMPLE_PATH="/home/samir/projects/mani/_example"
OUTPUT_FILE="$MANI_PATH/res/output.json"
OUTPUT_GIF="$MANI_PATH/res/output.gif"

_init() {
  # cd into _example
  cd "$MANI_EXAMPLE_PATH"

  # remove previous artifacts
  rm "$OUTPUT_FILE" "$OUTPUT_GIF" -f

  # remove previously synced projects
  paths=$(mani list projects --no-borders --no-headers --headers=path)
  for p in ${paths[@]}; do
    if [[ "$p" != "$MANI_EXAMPLE_PATH" ]]; then
      rm "$p" -rf
    fi
  done
}

_simulate_commands() {
  # list of the commands we want to record
  local CMD='
    _mock() {
      # the | pv -qL 30 part is used to simulate typing
      echo "\$ $1" | pv -qL 40

      first_char=$(printf %.1s "$1")
      if test "$first_char" != "'#'"; then
        $1
      fi
    }

    clear
    export PS1="\$ "
    sleep 2s

    # 1. List all projects
    _mock "# List all projects"
    sleep 2s
    _mock "mani list projects"
    sleep 4s
    clear

    # 2. Sync all repositories
    _mock "# Clone all those repositories"
    sleep 2s
    _mock "mani sync"
    sleep 5s
    clear
    _mock "# lets run an ad-hoc command to list files in template-generator"
    _mock "mani exec ls --projects template-generator"
    sleep 5s
    clear

    # 3. List all tasks
    _mock "# List all tasks"
    sleep 2s
    _mock "mani list tasks"
    sleep 4s
    clear

    # 4. Describe a command
    _mock "# See what git-status does"
    sleep 2s
    _mock "mani describe tasks git-status"
    sleep 4s
    clear

    # 5. Run a command
    _mock "# Now run git-status on all projects with node tag"
    sleep 2s
    _mock "mani run git-status --tags node --output table"
    sleep 4s
    clear

    # 6. Run a command
    _mock "# Check some random git stats"
    sleep 2s
    _mock "mani run git-daily --tags node --describe=false --output table"
    sleep 4s
    clear

    # 7. Run command on node repositories
    _mock "# Create a branch on multiple repositories"
    sleep 2s
    _mock "mani run git-create branch=feat/some-feature --tags node"
    sleep 4s
    clear

    # 8. Run command on node repositories
    _mock "# Install packages in all node repositories"
    sleep 2s
    _mock "mani run npm-install"
    sleep 6s
    clear
  '

  asciinema rec -c "$CMD" --idle-time-limit 100 --title mani --quiet "$OUTPUT_FILE" &
  fg %1
}

_generate_gif() {
  cd "$MANI_PATH/res"

  # Convert to gif
  output_file=$(basename $OUTPUT_FILE)
  output_gif=$(basename $OUTPUT_GIF)
  docker run --rm -v "$PWD":/data asciinema/asciicast2gif -S 3 -h 30 "$output_file" "$output_gif"
}

_main() {
  _init
  _simulate_commands
  _generate_gif
}

_main
