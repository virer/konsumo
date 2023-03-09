#!/bin/bash
# export SECRET_KEY="MyVerySecretKey"
# export GOOGLE_CLIENT_ID="123"
# export GOOGLE_CLIENT_SECRET="456"

# podman build . -t konsumo

podman run --rm \
    -v ./app/:/app/  \
    --name konsumo --network host -it  \
    -e HOST=0.0.0.0  \
    -e GOOGLE_CLIENT_ID="$GOOGLE_CLIENT_ID"  \
    -e GOOGLE_CLIENT_SECRET="$GOOGLE_CLIENT_SECRET"  \
    -e SECRET_KEY="$SECRET_KEY" \
    -e DBHOST="127.0.0.1" \
    -e DBPORT="3306" \
    -e DBUSER="root" \
    -e DBPASS="password" \
    -e DBNAME="konsumo" \
    localhost/konsumo:latest $1

# EOF