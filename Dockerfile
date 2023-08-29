FROM python:3.10-slim

RUN useradd myapp && mkdir /data && chown myapp: /data

ENV PYTHONDONTWRITEBYTECODE 1 \
    PYTHONUNBUFFERED 1

ENV DBUSER=root \
    DBPASS=password \
    DBHOST=127.0.0.1 \
    DBNAME=konsumo \
    SSL_CRT=/ssl/cert.pem \
    SSL_KEY=/ssl/key.pem

EXPOSE 8080

COPY requirements.txt /

RUN pip install --no-cache-dir -r /requirements.txt && rm -f /requirements.txt

COPY app/ /app/

USER myapp
WORKDIR /app

CMD [ "gunicorn", "--config", "gunicorn.conf.py", "app:app" ]