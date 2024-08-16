#!/usr/bin/env bash
set -Eeu

export ROOT_DIR="$(git rev-parse --show-toplevel)"
source ${ROOT_DIR}/test/system/lib.sh

export KOSLI_HOST=http://localhost:8001
export KOSLI_API_TOKEN="95-IeGBfyKdTteLdKidiAnXk6uMmV6jTkGM9v3DEtrQ"

export KOSLI_ORG=system-tests-user-shared
export KOSLI_FLOW=simple-flow
export KOSLI_TRAIL=first


testSimple() {
    ${ROOT_DIR}/kosli create flow "${KOSLI_FLOW}" \
      --use-empty-template
    assertEquals 0 $?
    ${ROOT_DIR}/kosli begin trail "${KOSLI_TRAIL}"
    assertEquals 0 $?
}

testPRexample() {
    ${ROOT_DIR}/kosli attest pullrequest github \
      --name=outer.inner \
      --commit=$(git rev-parse HEAD) \
      --github-org=kosli-dev \
      --github-token=sdsdf \
      --repository="${ROOT_DIR}" \
      --dry-run
    assertEquals 0 $?
}

# Load shUnit2.
source ${ROOT_DIR}/test/system/shunit2