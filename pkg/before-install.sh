#!/bin/bash

if ! id dedis &>/dev/null; then
	echo "user dedis not found"
	exit 1
fi

# Ideally these checks should be replaced by checking for these packages...

BINARIES=("/home/dedis/twins/biobank" "/home/dedis/twins/researcher" "/home/dedis/odyssey/ledger/conode/run_nodes.sh" "/home/dedis/odyssey/domanager/app/app" "/home/dedis/gb-backend/gb-backend/build/libs/gb-backend-0.1-SNAPSHOT.war" "/home/dedis/indy-vdr/target/release/indy-vdr-proxy" "/home/dedis/keycloak/keycloak-9.0.3/bin/standalone.sh" "/home/dedis/minio/bin/minio" "/home/dedis/transmart-core/transmart-api-server/build/libs/transmart-api-server-17.2-SNAPSHOT.war")
DIRECTORIES=("/home/dedis/aries-mediator/MediatorAgent")
MISSING=0

for binary in ${BINARIES[@]}; do
	if [ ! -f "$binary" ]; then
		echo "Could not find $binary"
		MISSING=1
	fi
done

for directory in ${DIRECTORIES[@]}; do
	if [ ! -d "$directory" ]; then
		echo "Could not find $directory"
		MISSING=1
	fi
done

if [[ "$MISSING" -eq 1 ]]; then
	exit 1
fi
