import os

import click
from flask import Flask
from flask.cli import with_appcontext
from flask_sqlalchemy import SQLAlchemy

from flask_login import LoginManager, current_user, login_required, login_user, logout_user
from oauthlib.oauth2 import WebApplicationClient

login_manager = LoginManager()
db = SQLAlchemy()
client = None

def create_app(test_config=None):
    global client, login_manager
    """Create and configure an instance of the Flask application."""
    
    app = Flask(__name__, instance_relative_config=True, static_url_path='/konsumo/static')
    app.config['SEND_FILE_MAX_AGE_DEFAULT'] = 0
    app.config.from_pyfile("config.py", silent=False)
    
    # initialize Flask-SQLAlchemy and the init-db command
    db.init_app(app)
    app.cli.add_command(init_db_command)

    # # OAuth 2 client setup
    client = WebApplicationClient(app.config['GOOGLE_CLIENT_ID'])
    # # User session management setup
    # # https://flask-login.readthedocs.io/en/latest
    login_manager.init_app(app)

    # Import blueprint
    from konsumo import auth, site
    app.register_blueprint(auth.bp)
    app.register_blueprint(site.bp)

    # make "index" point at "/", which is handled by "konsumo.index"
    app.add_url_rule("/", endpoint="index")

    return app

def init_db():
    db.drop_all()
    db.create_all()

@click.command("init-db")
@with_appcontext
def init_db_command():
    """Clear existing data and create new tables."""
    init_db()
    click.echo("Initialized the database.")
