#!/bin/sh

# Inspired from Debian packages (e.g. /var/lib/dpkg/info/openssh-server.postinst)
# In case this system is running systemd, we make systemd reload the unit files
# to pick up changes.
if [ -d /run/systemd/system ] ; then
    su -c "systemctl --system daemon-reload >/dev/null || true" dedis
fi
