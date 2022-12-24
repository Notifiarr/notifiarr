#!/usr/bin/env bash
# This file builds a standard DMG installer for macOS.
# This only works on macOS.
###########################################

# If we are running in Travis, make a new keychain and import the certificate.
if [ -n "$TRAVIS_OS_NAME" ] || [ -n "$APPLE_SIGNING_KEY" ]; then
  KEYCHAIN="ios-build.keychain"

  echo "==> Creating new keychain: $KEYCHAIN"
  security create-keychain -p secret $KEYCHAIN

  echo "==> Importing certificate into ${KEYCHAIN}"
  if [ -f "apple.signing.key" ]; then
    security import apple.signing.key -P '' -f pkcs12 -k $KEYCHAIN -T /usr/bin/codesign
  else
    echo "${APPLE_SIGNING_KEY}" | base64 -d | security import /dev/stdin -P '' -f pkcs12 -k $KEYCHAIN -T /usr/bin/codesign
  fi

  echo "==> Unlocking keychain ${KEYCHAIN}"
  security unlock-keychain -p secret $KEYCHAIN

  echo "==> Increase keychain unlock timeout to 1 hour."
  security set-keychain-settings -lut 3600 $KEYCHAIN
  
  security set-key-partition-list -S apple-tool:,apple: -s -k secret $KEYCHAIN

  echo "==> Add keychain to keychain-list"
  security list-keychains -s $KEYCHAIN
fi

echo "==> Signing App."
gon init/macos/sign.json

# Creating non-notarized DMG.
mkdir -p release
hdiutil create release/Notifiarr.dmg -srcfolder Notifiarr.app -ov

echo "==> Notarizing DMG."
gon init/macos/notarize.json

echo "==> Finished."
exit 0
