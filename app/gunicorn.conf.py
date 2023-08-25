import os
import multiprocessing

host = os.getenv("HOST", "0.0.0.0")
port = os.getenv("PORT", "8080")
bind = "{}:{}".format(host, port)
log_config = "/app/logger.ini"
workers = multiprocessing.cpu_count() + 1

# certfile = '/etc/letsencrypt/live/example.com/fullchain.pem'
# keyfile = '/etc/letsencrypt/live/example.com/privkey.pem'
certfile= os.getenv("SSL_CRT", "/ssl/cert.pem")
keyfile = os.getenv("SSL_KEY", "/ssl/key.pem")

debug = True
