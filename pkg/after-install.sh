#!/bin/sh

# fix permissions
chown -R dedis:dedis /opt/twins
chmod 755 /opt/twins

DEDIS_UID=$(id -u dedis)

# Always ensure systemd --user session lingers even after user logs out
loginctl enable-linger dedis

su -c "XDG_RUNTIME_DIR=/run/user/${DEDIS_UID} systemctl --user daemon-reload" dedis
su -c "XDG_RUNTIME_DIR=/run/user/${DEDIS_UID} systemctl --user start twins.target" dedis
su -c "XDG_RUNTIME_DIR=/run/user/${DEDIS_UID} systemctl --user enable twins.target" dedis
su -c "XDG_RUNTIME_DIR=/run/user/${DEDIS_UID} systemctl --user start twins_demo_reset.service" dedis
su -c "XDG_RUNTIME_DIR=/run/user/${DEDIS_UID} systemctl --user enable twins_demo_reset.service" dedis
