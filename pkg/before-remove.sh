#!/bin/sh

DEDIS_UID=$(id -u dedis)

# stop service
su -c "XDG_RUNTIME_DIR=/run/user/${DEDIS_UID} systemctl --user stop twins_demo_reset.service" dedis
su -c "XDG_RUNTIME_DIR=/run/user/${DEDIS_UID} systemctl --user stop twins.target" dedis

# remove binary
rm -rf /opt/twins/bin
