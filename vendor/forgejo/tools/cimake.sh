#!/bin/bash

# GNU make wrapper for invocation from ci to generate grouping of output

# pipefail: fail if make fails, because sed will not
set -euo pipefail

script='
  /^ *Prerequisite .* is newer than target /d;
  /^ * File .* does not exist/d;
  /^ *Must remake target/ s:^ *Must remake target:\#\#[group]:;
  /^ *Successfully remade target/ s:^ *Successfully remade target file:\#\#[endgroup]:;
  '

echo 'NOTICE: This is post-processed make output with some messages suppressed'
echo

# -O is important to not mix up output for parallel makes
exec make --debug=b -O -j$(nproc) "$@" 2>&1 | sed -u -e "${script}"
