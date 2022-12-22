#!/usr/bin/env bash

## This doesn't work in CI/CD. Must be run on a real Mac with a real apple ID and password.

source settings.sh

# Initial r/w image.
echo "Creating pack.temp.dmg."
rm -f pack.temp.dmg
hdiutil create -srcfolder "${MACAPP}.app" -volname "${MACAPP}" -fs HFS+ \
      -fsargs "-c c=64,a=16,e=16" -format UDRW -size 200000k pack.temp.dmg

# Get device.
hdiutil detach "/Volumes/${MACAPP}"
sleep 1
echo "Attaching pack.temp.dmg."
device=$(hdiutil attach -readwrite -noverify -noautoopen "pack.temp.dmg" | \
         egrep '^/dev/' | sed 1q | awk '{print $1}')

# Create content.
sleep 1
echo "Copying background."
mkdir "/Volumes/${MACAPP}/.background"
cp -r init/macos/background.png "/Volumes/${MACAPP}/.background/${MACAPP}.png"

echo "Running AppleScript to build custom DMG."
echo '
   tell application "Finder"
     tell disk "'${MACAPP}'"
           open
           set current view of container window to icon view
           set toolbar visible of container window to false
           set statusbar visible of container window to false
           set the bounds of container window to {400, 100, 1320, 600}
           set theViewOptions to the icon view options of container window
           set arrangement of theViewOptions to not arranged
           set icon size of theViewOptions to 256
           set background picture of theViewOptions to file ".background:'${MACAPP}.png'"
           make new alias file at container window to POSIX file "/Applications" with properties {name:"Applications"}
           set position of item "'${MACAPP}.app'" of container window to {0, 0}
           set position of item "Applications" of container window to {600, 0}
           update without registering applications
           delay 1
           close
     end tell
   end tell
' | osascript

sleep 1
# Finalize.
echo "Finalizing DMG."
chmod -Rf go-w /Volumes/"${MACAPP}"

sleep 1
echo "Detaching ${device}."
hdiutil detach ${device}

sleep 1
echo "Converting DMG."
mkdir -p release
rm -f "release/${MACAPP}.dmg"
hdiutil convert "pack.temp.dmg" -format UDZO -imagekey zlib-level=9 -o "release/${MACAPP}.dmg"
rm -f pack.temp.dmg 

echo "Notarizing DMG."
gon init/macos/notarize.json
