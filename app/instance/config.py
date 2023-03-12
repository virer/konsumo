import os

# Configuration
SECRET_KEY = os.environ.get("SECRET_KEY")
GOOGLE_CLIENT_ID = os.environ.get("GOOGLE_CLIENT_ID", None)
GOOGLE_CLIENT_SECRET = os.environ.get("GOOGLE_CLIENT_SECRET", None)
GOOGLE_DISCOVERY_URL = ("https://accounts.google.com/.well-known/openid-configuration")


SQLALCHEMY_DATABASE_URI = 'mysql+pymysql://{}:{}@{}/{}'.format(
    os.environ.get('DBUSER', 'root'),
    os.environ.get('DBPASS', 'password'),
    os.environ.get('DBHOST', '127.0.0.1'),
    os.environ.get('DBNAME', 'konsumo')
)
