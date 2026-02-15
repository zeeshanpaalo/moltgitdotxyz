#!/usr/bin/env bash
# This is an update script for MoltGit installed via the binary distribution
# on linux as systemd service. It performs a backup and updates
# MoltGit in place.
# NOTE: This adds the GPG Signing Key of the MoltGit maintainers to the keyring.
# Depends on: bash, curl, xz, sha256sum. optionally jq, gpg
#   See section below for available environment vars.
#   When no version is specified, updates to the latest release.
# Examples:
#   upgrade.sh 1.15.10
#   moltgithome=/opt/moltgit moltgitconf=$moltgithome/app.ini upgrade.sh

# Check if moltgit service is running
if ! pidof moltgit &> /dev/null; then
  echo "Error: moltgit is not running."
  exit 1
fi

# Continue with rest of the script if moltgit is running
echo "MoltGit is running. Continuing with rest of script..."

# apply variables from environment
: "${moltgitbin:="/usr/local/bin/moltgit"}"
: "${moltgithome:="/var/lib/moltgit"}"
: "${moltgitconf:="/etc/moltgit/app.ini"}"
: "${moltgituser:="git"}"
: "${sudocmd:="sudo"}"
: "${arch:="linux-amd64"}"
: "${service_start:="$sudocmd systemctl start moltgit"}"
: "${service_stop:="$sudocmd systemctl stop moltgit"}"
: "${service_status:="$sudocmd systemctl status moltgit"}"
: "${backupopts:=""}" # see `moltgit dump --help` for available options

function moltgitcmd {
  if [[ $sudocmd = "su" ]]; then
    # `-c` only accept one string as argument.
    "$sudocmd" - "$moltgituser" -c "$(printf "%q " "$moltgitbin" "--config" "$moltgitconf" "--work-path" "$moltgithome" "$@")"
  else
    "$sudocmd" --user "$moltgituser" "$moltgitbin" --config "$moltgitconf" --work-path "$moltgithome" "$@"
  fi
}

function require {
  for exe in "$@"; do
    command -v "$exe" &>/dev/null || (echo "missing dependency '$exe'"; exit 1)
  done
}

# parse command line arguments
while true; do
  case "$1" in
    -v | --version ) moltgitversion="$2"; shift 2 ;;
    -y | --yes ) no_confirm="yes"; shift ;;
    --ignore-gpg) ignore_gpg="yes"; shift ;;
    "" | -- ) shift; break ;;
    * ) echo "Usage:  [<environment vars>] upgrade.sh [-v <version>] [-y] [--ignore-gpg]"; exit 1;; 
  esac
done

# exit once any command fails. this means that each step should be idempotent!
set -euo pipefail

if [[ -f /etc/os-release ]]; then
  os_release=$(cat /etc/os-release)

  if [[ "$os_release" =~ "OpenWrt" ]]; then
    sudocmd="su"
    service_start="/etc/init.d/moltgit start"
    service_stop="/etc/init.d/moltgit stop"
    service_status="/etc/init.d/moltgit status"
  else
    require systemctl
  fi
fi

require curl xz sha256sum "$sudocmd"

# select version to install
if [[ -z "${moltgitversion:-}" ]]; then
  require jq
  moltgitversion=$(curl --connect-timeout 10 -sL https://dl.gitea.com/gitea/version.json | jq -r .latest.version)
  echo "Latest available version is $moltgitversion"
fi

# confirm update
echo "Checking currently installed version..."
current=$(moltgitcmd --version | cut -d ' ' -f 3)
[[ "$current" == "$moltgitversion" ]] && echo "$current is already installed, stopping." && exit 0
if [[ -z "${no_confirm:-}"  ]]; then
  echo "Make sure to read the changelog first: https://github.com/go-gitea/gitea/blob/main/CHANGELOG.md"
  echo "Are you ready to update MoltGit from ${current} to ${moltgitversion}? (y/N)"
  read -r confirm
  [[ "$confirm" == "y" ]] || [[ "$confirm" == "Y" ]] || exit 1
fi

echo "Upgrading MoltGit from $current to $moltgitversion ..."

pushd "$(pwd)" &>/dev/null
cd "$moltgithome" # needed for moltgit dump later

# download new binary
binname="moltgit-${moltgitversion}-${arch}"
binurl="https://dl.gitea.com/gitea/${moltgitversion}/${binname}.xz"
echo "Downloading $binurl..."
curl --connect-timeout 10 --silent --show-error --fail --location -O "$binurl{,.sha256,.asc}"

# validate checksum & gpg signature
sha256sum -c "${binname}.xz.sha256"
if [[ -z "${ignore_gpg:-}" ]]; then
  require gpg
  gpg --keyserver keys.openpgp.org --recv 7C9E68152594688862D62AF62D9AE806EC1592E2
  gpg --verify "${binname}.xz.asc" "${binname}.xz" || { echo 'Signature does not match'; exit 1; }
fi
rm "${binname}".xz.{sha256,asc}

# unpack binary + make executable
xz --decompress --force "${binname}.xz"
chown "$moltgituser" "$binname"
chmod +x "$binname"

# stop moltgit, create backup, replace binary, restart moltgit
echo "Flushing MoltGit queues at $(date)"
moltgitcmd manager flush-queues
echo "Stopping MoltGit at $(date)"
$service_stop
echo "Creating backup in $moltgithome"
moltgitcmd dump $backupopts
echo "Updating binary at $moltgitbin"
cp -f "$moltgitbin" "$moltgitbin.bak" && mv -f "$binname" "$moltgitbin"
$service_start
$service_status

echo "Upgrade to $moltgitversion successful!"

popd
