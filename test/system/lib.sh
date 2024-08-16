
exit_non_zero_unless_installed() {
  for dependent in "$@"
  do
    if ! installed "${dependent}" ; then
      stderr "${dependent} is not installed"
      exit 42
    fi
  done
}

installed() {
  local -r dependent="${1}"
  if hash "${dependent}" 2> /dev/null; then
    true
  else
    false
  fi
}

stderr() {
  local -r message="${1}"
  >&2 echo "ERROR: ${message}"
}
