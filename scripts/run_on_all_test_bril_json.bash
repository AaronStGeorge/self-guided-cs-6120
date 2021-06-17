#!/bin/bash
set -euo pipefail

echo "Note: This only asserts that the program will run with non zero exit code on the
      inputs naively gathered from bril test directory. Not a particularly strong test."
echo

BRIL_TEST_DIR=${BRIL_TEST_DIR:-~/Dev/misc/bril/test}

# TODO document what this find command does
for i in $(find $BRIL_TEST_DIR -type f -name "*.bril" ! -name "spec*" ! -name "ssa*"); do
  if bril2json <"$i" | "./$1" >/dev/null; then
    echo "ok - $i"
  else
    echo "fail - $i"
    exit 1
  fi
done
