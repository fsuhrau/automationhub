#!/usr/bin/env bash
# Usage: scripts/go-cover-diff.sh <base> <coverprofile> [dir...]
#
# The optional dirs limit the diff to Go files below them (default: whole repo).
#
# Reports every changed or added executable Go line that the coverprofile shows
# as not covered. The working tree (staged, unstaged and untracked files) is
# diffed against <base>. Test files (_test.go) and generated *.pb.go are ignored.
#
# A changed file absent from the profile is only reported if it declares a
# function; files without statements (types, consts, docs) have no profile blocks.
#
# Output: one "file:line: <reason>" per finding; exit 1 if there are any.
# For tests, GO_COVER_DIFF_FILE may name a prepared `git diff -U0` output that is
# used instead of running git.
set -euo pipefail

MODULE_PREFIX="github.com/fsuhrau/automationhub/"

usage() {
  echo "usage: $0 <base> <coverprofile> [dir...]" >&2
  exit 2
}

# Prints one git pathspec per Go file glob: "<dir>/*.go" per dir, else "*.go".
pathspecs() {
  [[ $# -eq 0 ]] && set -- .
  local dir
  for dir in "$@"; do
    [[ "$dir" == . ]] && echo '*.go' || echo "${dir%/}/*.go"
  done
}

# Prints the unified diff (zero context) of the Go files in $2.. against $1.
raw_diff() {
  if [[ -n "${GO_COVER_DIFF_FILE:-}" ]]; then
    cat "$GO_COVER_DIFF_FILE"
    return
  fi
  local base="$1"
  shift
  mapfile -t specs < <(pathspecs "$@")
  git diff -U0 "$base" -- "${specs[@]}"
  untracked_diff "${specs[@]}"
}

# Diffs every untracked Go file matching the pathspecs against /dev/null so its
# lines count as added.
untracked_diff() {
  local file
  while IFS= read -r file; do
    git diff --no-index -U0 -- /dev/null "$file" || true
  done < <(git ls-files --others --exclude-standard -- "$@")
}

# Turns a diff into "file\034line\034text" (\034 as separator, since text may
# contain tabs) for every added line of a non-test, non-generated Go file.
changed_lines() {
  awk '
    /^\+\+\+ / { file = ($2 == "/dev/null") ? "" : substr($2, 3); next }
    /^@@/ {
      split($3, plus, ",")
      line = substr(plus[1], 2) + 0
      next
    }
    /^\+/ && file ~ /\.go$/ && file !~ /(_test|\.pb)\.go$/ {
      printf "%s\034%d\034%s\n", file, line, substr($0, 2)
    }
    /^\+/ { line++ }
  '
}

# Turns a coverprofile into "file start end count" per block, with repo paths.
profile_blocks() {
  awk -v prefix="$MODULE_PREFIX" '
    NR == 1 && /^mode:/ { next }
    {
      split($1, loc, ":")
      split(loc[2], range, ",")
      split(range[1], from, ".")
      split(range[2], to, ".")
      file = loc[1]
      if (index(file, prefix) == 1) file = substr(file, length(prefix) + 1)
      print file, from[1], to[1], $3
    }
  ' "$1"
}

# Reads blocks (file $1) and changed lines (stdin); prints the uncovered ones.
report_uncovered() {
  awk -F"\034" -v blocksFile="$1" '
    function trivial(text) {
      gsub(/[ \t]/, "", text)
      return text == "" || text ~ /^\/\// || text ~ /^\/\*/ || text ~ /^[{}()\[\],]+$/
    }
    function hasFunc(file,    line, found) {
      while ((getline line < file) > 0) if (line ~ /^func /) { found = 1; break }
      close(file)
      return found
    }
    function state(file, line,    i, b, zero, hit) {
      for (i = 1; i <= count[file]; i++) {
        split(blocks[file, i], b, " ")
        if (line < b[1] || line > b[2]) continue
        if (b[3] > 0) hit = 1; else zero = 1
      }
      return hit ? "covered" : (zero ? "uncovered" : "none")
    }
    BEGIN {
      while ((getline row < blocksFile) > 0) {
        split(row, f, " ")
        blocks[f[1], ++count[f[1]]] = f[2] " " f[3] " " f[4]
      }
    }
    trivial($3) { next }
    !($1 in count) && !hasFunc($1) { next }
    !($1 in count) {
      if (!($1 in missing)) print $1 ":" $2 ": file missing from coverage profile"
      missing[$1] = 1; bad = 1
      next
    }
    state($1, $2) == "uncovered" { print $1 ":" $2 ": not covered"; bad = 1 }
    END { exit bad }
  '
}

main() {
  [[ $# -ge 2 ]] || usage
  local profile="$2"
  raw_diff "$1" "${@:3}" | changed_lines | report_uncovered <(profile_blocks "$profile")
}

main "$@"
