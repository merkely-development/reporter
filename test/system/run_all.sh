#!/usr/bin/env bash
set -Eeu

export ROOT_DIR="$(git rev-parse --show-toplevel)"

for file in ${ROOT_DIR}/test/system/test_*.sh; do
  echo "Running ${file}"
  ${file}
done