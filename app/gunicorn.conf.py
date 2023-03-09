import os
import multiprocessing

host = os.getenv("HOST", "0.0.0.0")
port = os.getenv("PORT", "8080")
bind = "{}:{}".format(host, port)
log_config = "/app/logger.ini"
workers = multiprocessing.cpu_count() * 2 + 1