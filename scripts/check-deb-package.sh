#!/usr/bin/env bash
set -euo pipefail

package=$(realpath "$1")
test "$(dpkg-deb -f "$package" Package)" = fittrack
profile=$(mktemp -d)
export XDG_DATA_HOME="$profile"
db="$profile/FitTrack/fittrack.sqlite3"

launch() {
  # Isolate the native WebView from the network but keep its loopback available.
  local status=0
  sudo unshare --net bash -c '
    ip link set lo up
    exec runuser -u "$1" --preserve-environment -- dbus-run-session -- xvfb-run -a timeout 12s /usr/local/bin/fittrack
  ' _ "$(id -un)" || status=$?
  test "$status" -eq 124
  test -s "$db"
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
