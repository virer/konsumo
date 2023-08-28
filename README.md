# Konsumo

## a little home energy consumption chart OpenSource projet

This project is based on Python, Flask, ApexCharts (apexcharts.com), SQLAlchemy, Pandas

### First time, run the following command to initialize de DB

```console
$ podman exec -it konsumo bash
...
$ flask init-db
```

You will need to generate your SSL certificate, here is a quick self-signed command example :
```console
openssl req -x509 -newkey rsa:4096 -keyout key.pem -nodes -out cert.pem -sha256 -days 3650 -subj='/CN=127.0.0.1/'
```

### Badges
![CI-PROD workflow](https://github.com/github/docs/actions/workflows/main.yml/badge.svg)    [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=virer_konsumo&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=virer_konsumo)

### Source of inspiration for the login part:
Google Login tutorial https://realpython.com/flask-google-login/
