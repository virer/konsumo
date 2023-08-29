#!/bin/bash
# export MARIADB_ROOT_PASSWORD="MyVerySecretPassword"

podman run --rm -d --name mariadb -v /data/mariadb/:/var/lib/mysql/ --network host \
    -e MARIADB_ROOT_PASSWORD="$MARIADB_ROOT_PASSWORD" \
    -e MARIADB_DATABASE="konsumo" \
    -e MARIADB_USER="konsumo" \
    -e MARIADB_PASSWORD="password" \
    mariadb:10.11 

# EOF