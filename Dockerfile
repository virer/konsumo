FROM python:3.10-slim

RUN useradd myapp && mkdir /data && chown myapp: /data

ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1
EXPOSE 8080

COPY requirements.txt /

RUN pip install --no-cache-dir -r /requirements.txt

COPY . /

USER myapp
WORKDIR /app

CMD [ "gunicorn", "--config", "gunicorn.conf.py", "app:app" ]