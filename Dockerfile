FROM python:3.10-slim

RUN useradd myapp && mkdir /data && chown myapp: /data

ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1
EXPOSE 8080

COPY requirements.txt /

RUN apt-get update && apt-get install -y gcc python3-dev && pip install --no-cache-dir -r /requirements.txt

# RUN apk add --no-cache mariadb-connector-c-dev
# RUN apk add --no-cache --virtual .build-deps build-base gcc musl-dev \
#     && pip install --no-cache-dir -r /requirements.txt \
#     && apk del .build-deps build-base gcc musl-dev

COPY . /

USER myapp
WORKDIR /app

# DEV MODE :
CMD [ "python", "konsumo.py" ]
# PROD MODE: but it require an SSL endpoint in front
# CMD [ "gunicorn", "--config", "gunicorn.conf.py", "konsumo:app" ]