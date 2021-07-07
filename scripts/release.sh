#!/bin/bash

set -euo pipefail

# Get latest version changes only
sed '0,/## v/d;/## v/Q' CHANGELOG.md | tail -n +2 | head -n-1 > release-changelog.md
