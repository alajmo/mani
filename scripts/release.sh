#!/bin/bash

set -euo pipefail

sed -n -e '/^## v/,/^## v/p' CHANGELOG.md | head -n -2 | tail -n +3 > release-changelog.md
