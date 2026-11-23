#!/usr/bin/env bash
set -euo pipefail

package=$(realpath "$1")
test "$(dpkg-deb -f "$package" Package)" = fittrack
profile=$(mktemp -d)
db="$profile/FitTrack/fittrack.sqlite3"

launch() {
  local status=0
  local webkit_env=()
  if [[ ${GITHUB_ACTIONS:-} == true ]]; then
    # The hosted runner blocks WebKit's nested user namespace.
    webkit_env+=(WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS=1)
  fi
  # Keep the test app offline and set its data path after dropping root privileges.
  sudo unshare --net bash -ec '
    ip link set lo up
    user=$1
    data_home=$2
    shift 2
    exec runuser -u "$user" -- env "XDG_DATA_HOME=$data_home" "$@" \
      dbus-run-session -- xvfb-run -a timeout 12s /usr/local/bin/fittrack
  ' _ "$(id -un)" "$profile" "${webkit_env[@]}" || status=$?
  if [[ $status != 124 || ! -s $db ]]; then
    echo "FitTrack offline launch failed (exit $status, expected timeout and database at $db)." >&2
    return 1
  fi
}

trap 'sudo dpkg -r fittrack >/dev/null; rm -rf -- "$profile"' EXIT
sudo dpkg -i "$package"
launch
node - "$db" <<'NODE'
const { DatabaseSync } = require('node:sqlite')
const db = new DatabaseSync(process.argv[2])
db.exec(`
  INSERT INTO app_settings VALUES (1, 'Package check', 'kg', 'system', 0, 0, 0, 0);
  INSERT INTO weight_measurements (measured_on, weight_kg) VALUES ('2026-11-23', 80);
`)
db.close()
NODE
before=$(sha256sum "$db")
sudo dpkg -r fittrack
test ! -e /usr/local/bin/fittrack
test "$(sha256sum "$db")" = "$before"
sudo dpkg -i "$package"
launch
test "$(sha256sum "$db")" = "$before"
echo 'Debian install, offline launch, removal and reinstall passed.'
