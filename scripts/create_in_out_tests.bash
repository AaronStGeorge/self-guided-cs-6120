#!/bin/bash
set -euo pipefail
# This script creates bril tests for in-out program. That program tests
# serialization invariants. Output should be the same as input after being run
# through out serialization code.

BRIL_TEST_DIR=${BRIL_TEST_DIR:-~/Dev/misc/bril/test}
TARGET_DIR=../test/in-out

rm -rf $TARGET_DIR

mkdir -p $TARGET_DIR
cd $TARGET_DIR

COUNTER=0
echo "Copying bril files from $BRIL_TEST_DIR to $TARGET_DIR"
for FILE in $(find "$BRIL_TEST_DIR" -type f -name "*.bril" ! -name "spec*" ! -name "ssa*"); do
  # Copy only if they don't have any errors when run through bril2json
  if bril2json < "$FILE" > /dev/null 2> /dev/null; then
    cp "$FILE" .
    (( COUNTER++ )) || true
    echo "$COUNTER - copied $FILE"
  fi
done

echo "command = \"bril2json < {filename} | jq\"" > turnt.toml

echo "Saving output of bril2json"
turnt ./*.bril --save || true

echo "Writing real turnt.toml"
echo "command = \"bril2json < {filename} | ../../bin/in-out | jq\"" > turnt.toml
turnt ./*.bril || true
