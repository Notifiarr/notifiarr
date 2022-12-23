#!/usr/bin/env bash

## This doesn't work in CI/CD because of the AppleScript. 
# Must be run on a real Mac with a real apple ID and password.
# AC_PASSWORD and AC_USERNAME variables are required to be set.
# The only part that doesn't work in CI/CD currently is the applescript.

# Download latest release.
echo "==> Getting latest release."
URL=$(curl -s https://api.github.com/repos/Notifiarr/notifiarr/releases/latest | \
    jq -r '.assets[] | select(.name == "Notifiarr.dmg") | .browser_download_url')
echo "==> Downloading: $URL"
curl -sSLo /tmp/Notifiarr.dmg "$URL"
echo "==> Mounting Notifiarr.dmg to /Volumes/NotifiarrRelease"
hdiutil attach -readonly -mountpoint /Volumes/NotifiarrRelease /tmp/Notifiarr.dmg

# Create r/w image with latest release app as source.
echo "==> Creating intermediate image: pack.temp.dmg."
rm -f pack.temp.dmg
hdiutil create -srcfolder "/Volumes/NotifiarrRelease/Notifiarr.app" -volname "Notifiarr" -fs HFS+ \
      -fsargs "-c c=64,a=16,e=16" -format UDRW -size 200000k pack.temp.dmg

echo "==> Unmounting /Volumes/NotifiarrRelease and /Volumes/NotifiarrIntermediate (may not be mounted)."
hdiutil detach "/Volumes/NotifiarrIntermediate"
hdiutil detach "/Volumes/NotifiarrRelease"
sleep 1

echo "==> Mounting pack.temp.dmg to /Volumes/NotifiarrIntermediate"
hdiutil attach -mountpoint /Volumes/NotifiarrIntermediate -readwrite -noverify -noautoopen "pack.temp.dmg" | \
         egrep '^/dev/' | sed 1q | awk '{print $1}'

# Create content.
sleep 1
echo "==> Copying background image."
mkdir "/Volumes/NotifiarrIntermediate/.background"
cp -r init/macos/background.png "/Volumes/NotifiarrIntermediate/.background/Notifiarr.png"

echo "==> Running AppleScript to build custom DMG."
echo '
   tell application "Finder"
     tell disk "'NotifiarrIntermediate'"
           open
           set current view of container window to icon view
           set toolbar visible of container window to false
           set statusbar visible of container window to false
           set the bounds of container window to {400, 100, 1320, 600}
           set theViewOptions to the icon view options of container window
           set arrangement of theViewOptions to not arranged
           set icon size of theViewOptions to 256
           set background picture of theViewOptions to file ".background:'Notifiarr.png'"
           make new alias file at container window to POSIX file "/Applications" with properties {name:"Applications"}
           set position of item "'Notifiarr.app'" of container window to {0, 0}
           set position of item "Applications" of container window to {600, 0}
           update without registering applications
           delay 1
           close
     end tell
   end tell
' | osascript

sleep 1
# Finalize.
echo "==> Finalizing DMG."
chmod -Rf go-w /Volumes/NotifiarrIntermediate

sleep 1
echo "==> Unmounting /Volumes/NotifiarrIntermediate."
hdiutil detach /Volumes/NotifiarrIntermediate

sleep 1
echo "==> Converting DMG to compressed read only."
mkdir -p release
rm -f "release/Notifiarr.dmg"
hdiutil convert "pack.temp.dmg" -format UDZO -imagekey zlib-level=9 -o "release/Notifiarr.dmg"
rm -f pack.temp.dmg 

echo "==> Notarizing DMG."
gon init/macos/notarize.json
echo "==> Finished!"
ls -l release/Notifiarr.dmg