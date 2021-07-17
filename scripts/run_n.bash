#!/bin/bash
set -euo pipefail
# this script runs the given command n times and will fail with a non-zero exit code if any invocation fails


N=$1
FILE=$2

for ((i=1; i<=N; i++)); do
  echo "$i"
  if ! turnt "$FILE" > /dev/null; then
    echo "fail!"
    exit 1
  fi
done
echo "success!"

