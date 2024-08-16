#!/usr/bin/env bash
set -Eeu

export ROOT_DIR="$(git rev-parse --show-toplevel)"
source ${ROOT_DIR}/test/system/lib.sh

testSimple() {
    assertEquals 1 1
}



# Load shUnit2.
source ${ROOT_DIR}/test/system/shunit2