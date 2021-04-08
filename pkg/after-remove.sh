#!/bin/sh

DEDIS_UID=$(id -u dedis)

# Inspired from Debian packages (e.g. /var/lib/dpkg/info/openssh-server.postinst)
# In case this system is running systemd, we make systemd reload the unit files
# to pick up changes.
if [ -f /run/systemd/users/$DEDIS_UID ] ; then
    su -c "XDG_RUNTIME_DIR=/run/user/${DEDIS_UID} systemctl --user daemon-reload >/dev/null || true" dedis
fi
