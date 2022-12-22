#!/usr/bin/env bash
# This file builds a beautiful DMG installer for macOS.
# This only works on macOS.
###########################################

source settings.sh

[ -n "$MACAPP" ] || exit 1

# If we are running in Travis, make a new keychain and import the certificate.
if [ -n "$TRAVIS_OS_NAME" ]; then
  KEYCHAIN="ios-build.keychain"

  echo "Creating new keychain: $KEYCHAIN"
  security create-keychain -p secret $KEYCHAIN

  echo "Importing certificate into ${KEYCHAIN}"
  security import apple.signing.key -P '' -f pkcs12 -k $KEYCHAIN -T /usr/bin/codesign

  echo "Unlocking keychain ${KEYCHAIN}"
  security unlock-keychain -p secret $KEYCHAIN

  echo "Increase keychain unlock timeout to 1 hour."
  security set-keychain-settings -lut 3600 $KEYCHAIN
  
  security set-key-partition-list -S apple-tool:,apple: -s -k secret $KEYCHAIN

  echo "Add keychain to keychain-list"
  security list-keychains -s $KEYCHAIN
fi

echo "Signing App."
gon init/macos/sign.json

# Creating non-notarized DMG.
mkdir -p release
hdiutil create release/${MACAPP}.dmg -srcfolder ${MACAPP}.app -ov

echo "Notarizing DMG."
gon init/macos/notarize.json