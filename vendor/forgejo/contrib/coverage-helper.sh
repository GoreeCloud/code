#!/bin/bash

set -e
#set -x
PS4='${BASH_SOURCE[0]}:$LINENO: ${FUNCNAME[0]}:  '

#
# Those must be explicitly required and are excluded from the full list of packages because they
# would interfere with the testing fixtures.
#
excluded+='forgejo.org/models/gitea_migrations|'                # must be run before database specific tests
excluded+='forgejo.org/models/forgejo_migrations_legacy|'        # must be run before database specific tests
excluded+='forgejo.org/tests/integration/migration-test|' # must be run before database specific tests
excluded+='forgejo.org/tests|'                            # only tests, no coverage to get there
excluded+='forgejo.org/tests/e2e|'                        # JavaScript is not in scope here and if it adds coverage it should not be counted
excluded+='FAKETERMINATOR'                                # do not modify

: ${COVERAGEDIR:=$(pwd)/coverage/data}
: ${GO:=$(go env GOROOT)/bin/go}

DEFAULT_TEST_PACKAGES=$($GO list ./... | grep -E -v "$excluded")

COVERED_PACKAGES=$($GO list ./...)
COVERED_PACKAGES=$(echo $COVERED_PACKAGES | sed -e 's/ /,/g')

function run_test() {
  local package="$1"
  if echo "$package" | grep --quiet --fixed-string ".."; then
    echo "$package contains a suspicious .."
    return 1
  fi

  local coverage="$COVERAGEDIR/$COVERAGE_TEST_DATABASE/$package"
  rm -fr $coverage
  mkdir -p $coverage

  #
  # -race cannot be used because it requires -covermode atomic which is
  # different from the end-to-end tests and would cause issues wen merging
  #
  $GO test -timeout=20m -tags='sqlite sqlite_unlock_notify' -cover $package -coverpkg $COVERED_PACKAGES $COVERAGE_TEST_ARGS -args -test.gocoverdir=$coverage |& grep -v 'warning: no packages being tested depend on matches for pattern'
}

function test_packages() {
  for package in ${@:-$DEFAULT_TEST_PACKAGES}; do
    run_test $package
  done
}

"$@"
