#!/bin/bash
# export SECRET_KEY="MyVerySecretKey"
# export GOOGLE_CLIENT_ID="123"
# export GOOGLE_CLIENT_SECRET="456"

# podman build . -t konsumo

podman run --rm $KONSUMO_DEV -v /ssl:/ssl \
    --name konsumo \
    --network host -it \
    -e HOST=0.0.0.0  \
    -e PORT=8080  \
    -e GOOGLE_CLIENT_ID="$GOOGLE_CLIENT_ID"  \
    -e GOOGLE_CLIENT_SECRET="$GOOGLE_CLIENT_SECRET"  \
    -e SECRET_KEY="$SECRET_KEY" \
    -e DBHOST="127.0.0.1" \
    -e DBPORT="3306" \
    -e DBUSER="root" \
    -e DBPASS="password" \
    -e DBNAME="konsumo" \
    -e SSL_CRT="/ssl/cert.pem" \
    -e SSL_KEY="/ssl/key.pem" \
    docker.io/scaps/konsumo:latest $1 $2

# EOF