#!/bin/bash

DIRNAME=$( dirname -- "$0" );

mysql -u root --password="$MARIADB_ROOT_PASSWORD" -e "CREATE DATABASE konsumo"
mysql -u root --password="$MARIADB_ROOT_PASSWORD" konsumo -c < $DIRNAME/schema.sql

# EOF