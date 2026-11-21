#!/usr/bin/env bash
# Usage: scripts/go-file-size.sh [dir] [max-lines]
#
# Fails when a Go file below dir (default: the working directory) has more than
# max-lines lines (default: 200). Tests (_test.go), generated *.pb.go files,
# keycodes.go (a data table) and the local, git-ignored out/ folder are exempt,
# as are hidden folders, frontend/ and node_modules/.
#
# Output: one "file: N lines (max M)" per offender; exit 1 if there are any.
set -euo pipefail

dir="${1:-.}"
max="${2:-200}"

[[ "$max" =~ ^[0-9]+$ ]] || { echo "usage: $0 [dir] [max-lines]" >&2; exit 2; }
[[ -d "$dir" ]] || { echo "$0: $dir is not a directory" >&2; exit 2; }

cd "$dir"
status=0
while IFS= read -r file; do
  lines=$(wc -l <"$file")
  if ((lines > max)); then
    echo "${file#./}: $lines lines (max $max)"
    status=1
  fi
done < <(find . -mindepth 1 \
  \( -name '.*' -o -name node_modules -o -name frontend -o -name out \) -prune -o \
  -name '*.go' ! -name '*_test.go' ! -name '*.pb.go' ! -name keycodes.go -print | sort)
exit "$status"
