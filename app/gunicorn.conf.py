import os
import multiprocessing

host = os.getenv("HOST", "0.0.0.0")
port = os.getenv("PORT", "8080")
bind = "{}:{}".format(host, port)

accesslog = '/dev/null' # possible value '/dev/stdout' or /dev/null
loglevel = 'debug' # possible value : info, error, debug
access_log_format = '%(h)s %(l)s %(u)s %(t)s "%(r)s" %(s)s %(b)s "%(f)s" "%(a)s"'

workers = multiprocessing.cpu_count() + 1

# Example of usage
# certfile = '/etc/letsencrypt/live/example.com/fullchain.pem'
# keyfile = '/etc/letsencrypt/live/example.com/privkey.pem'
certfile= os.getenv("SSL_CRT", "/ssl/cert.pem")
keyfile = os.getenv("SSL_KEY", "/ssl/key.pem")

debug = True
