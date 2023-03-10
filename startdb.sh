#!/bin/bash
# export MARIADB_ROOT_PASSWORD="MyVerySecretPassword"

podman run --rm -d --name mariadb -v ./sql/:/konsumo_sql/ -v /data/mariadb/:/var/lib/mysql/ --network host -e MARIADB_ROOT_PASSWORD="$MARIADB_ROOT_PASSWORD" mariadb:10.11 

# EOF