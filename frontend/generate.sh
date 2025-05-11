#!/bin/sh

# This is run by go generate from frontend.go.
# It generates the locales.go file.
# And builds the frontend using npm.

# Install dependencies.
npm install

# Build the frontend 'dist' directory.
npm run build

# Get all the locales. Remove the folder prefix and the '.json' suffix.
locales=$(ls -1 src/lib/locale/*.json | sed 's#src/lib/locale/\(.*\)\.json#\1#g')

# Create the locales.go file.
cat <<EOF > locales.go
package frontend

var langs = []string{
EOF

# Add each locale to the file.
for locale in $locales; do
  echo "    \"$locale\"," >> locales.go
done

# Close the file.
echo "}" >> locales.go
