from flask import Flask
from flask.cli import with_appcontext

from flask_sqlalchemy import SQLAlchemy
from flask_login import LoginManager
from oauthlib.oauth2 import WebApplicationClient
from flask_wtf import CSRFProtect
import click

login_manager = LoginManager()
db = SQLAlchemy()
client = None

def create_app():
    global client, login_manager
    """Create and configure an instance of the Flask application."""
    
    app = Flask(__name__, instance_relative_config=True, static_url_path='/konsumo/static')
    app.config['SEND_FILE_MAX_AGE_DEFAULT'] = 0

    # Debug SQL query :
    # app.config["SQLALCHEMY_ECHO"] = True
    app.config.from_pyfile("config.py", silent=False)
    
    csrf = CSRFProtect()
    csrf.init_app(app)

    # initialize Flask-SQLAlchemy and the init-db command
    db.init_app(app)
    app.cli.add_command(init_db_command)

    # # OAuth 2 client setup
    client = WebApplicationClient(app.config['GOOGLE_CLIENT_ID'])
    # # User session management setup
    # # https://flask-login.readthedocs.io/en/latest
    login_manager.init_app(app)

    # Import blueprint
    from konsumo import auth, site, api
    app.register_blueprint(auth.bp)
    app.register_blueprint(site.bp)
    app.register_blueprint(api.bp)

    # Exempt API from CSRF
    csrf.exempt(api.bp)

    # make "index" point at "/", which is handled by "konsumo.index"
    app.add_url_rule("/", endpoint="root")

    return app

def init_db():
    db.drop_all()
    db.create_all()
    db.session.commit()

@click.command("init-db")
@with_appcontext
def init_db_command():
    """Clear existing data and create new tables."""
    init_db()
    click.echo("Initialized the database.")
