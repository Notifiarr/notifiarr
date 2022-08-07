#!/bin/sh

# This file is used by aur, deb, rpm and BSD packages.
# FPM adds this as the after-install script.
# Edit this file as needed for your application.
# This file is only installed if FORMULA is set to service.

if [ -d /usr/local/etc/{{BINARY}} ]; then
  chown -R {{BINARY}}: /usr/local/etc/{{BINARY}}
fi

if [ -d /etc/{{BINARY}} ]; then
  chown -R {{BINARY}}: /etc/{{BINARY}}
fi

if [ -d /var/log/{{BINARY}} ]; then
  chown -R {{BINARY}}: /var/log/{{BINARY}}
fi

if [ -d /usr/local/var/log/{{BINARY}} ]; then
  chown -R {{BINARY}}: /usr/local/var/log/{{BINARY}}
fi

if [ -x "/bin/systemctl" ]; then
  # Reload and restart - this starts the application as user nobody.
  /bin/systemctl daemon-reload
  /bin/systemctl enable {{BINARY}}
  /bin/systemctl restart {{BINARY}}
fi
