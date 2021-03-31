#!/bin/sh

INSTALLDIR=/opt/twins

# stop service
su -c "systemctl --user stop twins_demo_reset.service" dedis
su -c "systemctl --user stop twins.target" dedis

# remove binary
rm -rf /opt/twins/bin
