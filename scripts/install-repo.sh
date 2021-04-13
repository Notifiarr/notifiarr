#!/bin/bash
#
# This script installs the Go-Lift APT and/or YUM repo(s) on a Linux system.
# When run on macOS it attempts to tap the golift homebrew repo.
# Optionally triggers a package install if $1 is non-empty.
#
### Install Notifiarr:
# curl -sL https://raw.githubusercontent.com/Notifiarr/notifiarr/main/scripts/install-repo.sh | sudo bash -s - notifiarr
#

APT=$(which apt)
YUM=$(which yum)
BREW=$(which brew)
PKG=$1

if [ -d /etc/apt/sources.list.d ] && [ "$APT" != "" ]; then
  curl -sL https://packagecloud.io/golift/pkgs/gpgkey | apt-key add -
  echo "deb https://packagecloud.io/golift/pkgs/ubuntu focal main" > /etc/apt/sources.list.d/golift.list
  apt update
  [ "$PKG" = "" ] || apt install $PKG
fi

if [ -d /etc/yum.repos.d ] && [ "$YUM" != "" ]; then
  cat <<EOF > /etc/yum.repos.d/golift.repo
[golift]
name=golift
baseurl=https://packagecloud.io/golift/pkgs/el/6/\$basearch
repo_gpgcheck=1
gpgcheck=1
enabled=1
gpgkey=https://packagecloud.io/golift/pkgs/gpgkey
       https://packagecloud.io/golift/pkgs/gpgkey/golift-pkgs-7F7791485BF8996D.pub.gpg
sslverify=1
sslcacert=/etc/pki/tls/certs/ca-bundle.crt
metadata_expire=300
EOF

  yum -q makecache -y --disablerepo='*' --enablerepo='golift'
  [ "$PKG" = "" ] || yum install $PKG
fi

if [ "$(uname -s 2>/dev/null)" = "Darwin" ] && [ "$BREW" != "" ]; then
  brew tap golift/mugs
  [ "$PKG" = "" ] || brew install $PKG
fi
