#!/bin/bash

set -eum

MANI_PATH=$(dirname $(dirname $(realpath "$0")))
MANI_EXAMPLE_PATH="/home/samir/projects/mani/_examples"
OUTPUT_FILE="$MANI_PATH/res/output.json"
OUTPUT_GIF="$MANI_PATH/res/output.gif"

_init() {
  # cd into _example
  cd "$MANI_EXAMPLE_PATH"

  # remove previous artifacts
  rm "$OUTPUT_FILE" "$OUTPUT_GIF" -f

  # remove previously synced projects
  paths=$(mani list projects --headers=path)
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
    sleep 1s
    _mock "mani list projects"
    sleep 2s
    clear

    # 2. Sync all repositories
    _mock "# Clone all repositories"
    sleep 1s
    _mock "mani sync"
    sleep 3s
    clear

    # 3. Run command
    _mock "# lets run an ad-hoc command to list files in template-generator"
    sleep 1s
    _mock "mani exec ls --projects template-generator"
    sleep 3s
    clear

    # 4. List all tasks
    _mock "# List all tasks"
    sleep 1s
    _mock "mani list tasks"
    sleep 3s
    clear

    # 5. Run a command
    _mock "# Now run git-status on all projects with node tag"
    sleep 1s
    _mock "mani run git-status --tags node --output table"
    sleep 3s
    clear

    # 6. Run a command
    _mock "# Check some random git stats for all projects"
    sleep 1s
    _mock "mani run git-overview --all --output table"
    sleep 3s
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
