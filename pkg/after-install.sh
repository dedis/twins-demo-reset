#!/bin/sh

# fix permissions
chown dedis:dedis /opt/twins
chmod 755 /opt/twins


su -c "systemctl --user daemon-reload" dedis
su -c "systemctl --user start twins.target" dedis
su -c "systemctl --user start twins_demo_reset.service" dedis
