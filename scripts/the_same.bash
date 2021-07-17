#!/bin/bash
set -euo pipefail
# this script runs the given command n times and will fail with a non-zero exit code if any invocation fails

make build

N=$1
FILE=$2

for ((i=1; i<=N; i++)); do
  echo "$i"
  A=$(turnt "$FILE" --diff | sed -n '/@@.*/,$p' || true)
  B=$(turnt "$FILE" --diff | sed -n '/@@.*/,$p' || true)
#  echo "$A"
#  echo "$B"
  if [[ "$A" != "$B" ]]; then
    echo "fail!"
    exit 1
  fi
done
echo "success!"
