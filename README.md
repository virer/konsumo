# Konsumo

## About
Konsumo is a little home energy consumption chart OpenSource project

This project is based on Python, Flask, ApexCharts (apexcharts.com), SQLAlchemy, Pandas

### Quick start
First time, your have to create the SSL certificate(self signed here)

You will also need to have SSL certificate.

So here is a quick self-signed command example :
```console
mkdir -p /opt/konsumo/ssl && openssl req -x509 -newkey rsa:4096 -keyout /opt/konsumo/ssl/key.pem -nodes -out /opt/konsumo/ssl/cert.pem -sha256 -days 3650 -subj='/CN=konsumo/'
```

Then start a mariadb container
```console
podman pull docker.io/mariadb:10.11
podman run --rm -d --name mariadb -v /data/mariadb/:/var/lib/mysql/ --network host -e MARIADB_ROOT_PASSWORD="MyVerySecretPassword" mariadb:10.11
```

also initialize the DB using the following:
```console
$ podman pull docker.io/scaps/konsumo:<tag>
$ podman run -it docker.io/scaps/konsumo:<tag> /usr/local/bin/flask init-db
```

### Parameters, default value and possible usage
HOST=0.0.0.0
PORT=8080
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
DBHOST=127.0.0.1
DBUSER=root
DBPASS=password
DBNAME=konsumo
SSL_CRT=/ssl/cert.pem
SSL_KEY=/ssl/key.pem

```console
podman run --rm -v /opt/konsumo/ssl:/ssl \
    --name konsumo \
    -e HOST=0.0.0.0  \
    -e PORT=8080  \
    -e GOOGLE_CLIENT_ID="$GOOGLE_CLIENT_ID"  \
    -e GOOGLE_CLIENT_SECRET="$GOOGLE_CLIENT_SECRET"  \
    -e SECRET_KEY="$SECRET_KEY" \
    -e DBHOST="mariadb" \
    -e DBPORT="3306" \
    -e DBUSER="konsumo" \
    -e DBPASS="Konsum0Secre7P4s5woRd" \
    -e DBNAME="konsumo" \
    -e SSL_CRT="/ssl/cert.pem" \
    -e SSL_KEY="/ssl/key.pem" \
    docker.io/scaps/konsumo:latest $1 $2
```

### Status / Badges

![CI-PROD workflow](https://github.com/virer/konsumo/actions/workflows/main.yml/badge.svg)    [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=virer_konsumo&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=virer_konsumo)

### Notes 

Source of inspiration for the login part:

Google Login tutorial https://realpython.com/flask-google-login/
