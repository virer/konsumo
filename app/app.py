#!/usr/bin/env python3
######################
# S.CAPS Mar 2023
# Konsumo
######################

# Logging level
DEBUG = True

from flask import Flask, render_template, redirect, url_for, request, jsonify, copy_current_request_context, abort, send_from_directory
from flask_login import LoginManager, current_user, login_required, login_user, logout_user, UserMixin
from flask_sqlalchemy import SQLAlchemy
from jinja2 import TemplateNotFound
from oauthlib.oauth2 import WebApplicationClient
from datetime import datetime, date, timedelta
import os, copy, json, requests, sqlalchemy
import pandas as pd

app = Flask(__name__, static_url_path='/konsumo/static')
app.secret_key = os.environ.get("SECRET_KEY") or os.urandom(24)
app.config['SEND_FILE_MAX_AGE_DEFAULT'] = 0

# Configuration
HOST = os.getenv("HOST", "127.0.0.1")
PORT = os.getenv("PORT", "8080")
GOOGLE_CLIENT_ID = os.environ.get("GOOGLE_CLIENT_ID", None)
GOOGLE_CLIENT_SECRET = os.environ.get("GOOGLE_CLIENT_SECRET", None)
GOOGLE_DISCOVERY_URL = ("https://accounts.google.com/.well-known/openid-configuration")

db = SQLAlchemy()
app.config["SQLALCHEMY_DATABASE_URI"] = 'mysql+pymysql://{}:{}@{}/{}'.format(
    os.getenv('DBUSER', 'root'),
    os.getenv('DBPASS', 'password'),
    os.getenv('DBHOST', '127.0.0.1'),
    os.getenv('DBNAME', 'konsumo')
    )
db.init_app(app)

def json_serial(obj):
    """JSON serializer for objects not serializable by default json code"""

    if isinstance(obj, (datetime, date)):
        return obj.isoformat()
    raise TypeError ("Type %s not serializable" % type(obj))
class UsrDB(db.Model):
    __tablename__ = "user"
    user_id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String)
    email = db.Column(db.String)
    profile_pic = db.Column(db.String)
    location = db.Column(db.String)


class DataDB(db.Model):
    __tablename__ = "user_data"
    id = db.Column(db.Integer, primary_key=True)
    date = db.Column(db.Date)
    type = db.Column(db.String)
    value1 = db.Column(db.Integer)
    value2 = db.Column(db.Integer)
    user_id = db.Column(db.Integer)

class User(UserMixin):
    
    def __init__(self, id_=None, name=None, email=None, profile_pic=None, location=""):
        self.id = id_
        self.name = name
        self.email = email
        self.profile_pic = profile_pic
        self.location = location

    def get(self, user_id):
        sql = sqlalchemy.select(
                    UsrDB.user_id,UsrDB.name,UsrDB.email,UsrDB.profile_pic,UsrDB.location
                ).where(UsrDB.user_id == user_id)
        
        try:
            row = db.session.execute( sql ) 
            user = list(row.fetchall())[0]
        except Exception as e:
            if DEBUG:
                print(e)
            return False
        
        user = User(
            id_=user[0], name=user[1], email=user[2], profile_pic=user[3], location=user[4]
        )
        return user

    def create(self, id_, name, email, profile_pic):
        db.session.execute(
            f"INSERT INTO user (user_id, name, email, profile_pic) "
             "VALUES (?, ?, ?, ?)",
             (id_, name, email, profile_pic),
        )
        db.session.commit()
        db.session.close()
    
    def set_location(self, user_id, location):
        db.session.execute( f"UPDATE user set location = %s WHERE user_id = %d", (location, user_id) )
        db.session.commit()
        db.session.close()

    def set_data(self, date, type, value1, value2, user_id):
        if len(value2) > 0 :
            db.session.execute( f"INSERT INTO user_data (date, type, value1, value2, user_id) VALUES (?, ?, ?, ?, ?)", (date, type, value1, value2, user_id, ) )
        db.session.execute( f"INSERT INTO user_data (date, type, value1, user_id) VALUES (?, ?, ?, ?)", (date, type, value1, user_id, ) )
        db.session.commit()
        db.session.close()
    
    def get_data1(self, user_id, type):
        sql = sqlalchemy.select(
                    DataDB.date, DataDB.value1
                ).order_by(DataDB.date).where(DataDB.type == type).where(DataDB.user_id == user_id)
        
        try:
            rows = db.session.execute( sql )
            return rows.fetchall()
        except Exception as e:
            if DEBUG:
                print(e)
            return False

    def db_close(self):
        db.session.close()


def get_google_provider_cfg():
    return requests.get(GOOGLE_DISCOVERY_URL).json()

# # OAuth 2 client setup
client = WebApplicationClient(GOOGLE_CLIENT_ID)

# # User session management setup
# # https://flask-login.readthedocs.io/en/latest
login_manager = LoginManager()
login_manager.init_app(app)

@app.after_request
def set_response_headers(response):
    response.headers['Cache-Control'] = 'no-cache, no-store, must-revalidate'
    response.headers['Pragma'] = 'no-cache'
    response.headers['Expires'] = '0'
    return response

# Flask-Login helper to retrieve a user from our db
@login_manager.user_loader
def load_user(user_id):
    user = User()
    return user.get(user_id)

@app.route("/konsumo/login")
def login():
    # Find out what URL to hit for Google login
    google_provider_cfg = get_google_provider_cfg()
    authorization_endpoint = google_provider_cfg["authorization_endpoint"]

    # Use library to construct the request for Google login and provide
    # scopes that let you retrieve user's profile from Google
    request_uri = client.prepare_request_uri(
        authorization_endpoint,
        redirect_uri=request.base_url + "/callback",
        scope=["openid", "email", "profile"],
    )
    return redirect(request_uri)

@app.route("/konsumo/login/callback")
def callback():
    # Get authorization code Google sent back to you
    code = request.args.get("code")

    # Find out what URL to hit to get tokens that allow you to ask for
    # things on behalf of a user
    google_provider_cfg = get_google_provider_cfg()
    token_endpoint = google_provider_cfg["token_endpoint"]


    # Prepare and send a request to get tokens! Yay tokens!
    token_url, headers, body = client.prepare_token_request(
        token_endpoint,
        authorization_response=request.url,
        redirect_url=request.base_url,
        code=code
    )
    token_response = requests.post(
        token_url,
        headers=headers,
        data=body,
        auth=(GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET),
    )

    # Parse the tokens!
    client.parse_request_body_response(json.dumps(token_response.json()))

    # hit the URL from Google that gives you the user's profile information,
    # including their Google profile image and email
    userinfo_endpoint = google_provider_cfg["userinfo_endpoint"]
    uri, headers, body = client.add_token(userinfo_endpoint)
    userinfo_response = requests.get(uri, headers=headers, data=body)

    # You want to make sure their email is verified.
    # The user authenticated with Google, authorized your
    # app, and now you've verified their email through Google!
    if userinfo_response.json().get("email_verified"):
        unique_id = userinfo_response.json()["sub"]
        users_email = userinfo_response.json()["email"]
        picture = userinfo_response.json()["picture"]
        users_name = userinfo_response.json()["given_name"]
    else:
        return "User email not available or not verified by Google.", 400

    # Create a user in your db with the information provided
    # by Google
    user = User(
        id_=unique_id, name=users_name, email=users_email, profile_pic=picture
    )
    if DEBUG:
        print("UserId {} logged in".format(unique_id))

    # Doesn't exist? Add it to the database.
    if not user.get(unique_id):
        user.create(unique_id, users_name, users_email, picture)

    # Begin user session by logging the user in
    login_user(user)

    # Send user back to homepage
    return redirect(url_for("profile"))

@app.route("/konsumo/logout")
def logout():
    try:
        logout_user()
    except:
        pass
    return redirect(url_for("profile"))

def deltadays(dat1, dat2):
    date_format = "%Y-%m-%d"
    a = datetime.strptime(dat1, date_format)
    b = datetime.strptime(dat2, date_format)
    delta = b - a
    return int(delta.days)

last_value = 0
def diff_m(v):
        global last_value
        ret = v - last_value
        last_value = v
        return ret

def construct_data():
    global last_value
    # get data here
    user = User()
    data = user.get_data1(current_user.id, 'electricity')
    
    last_value = list(data)[0][1]

    return [(k, diff_m(v)) for k,v in data ]

def present_data(chartid):
    heating_period={ "start":"09", "end":"05" }

    data = construct_data()

    # transform here
    if DEBUG:
        print("DATA:**")
        print(data)

    if chartid == "current":
        fields = ['x', 'y']
        dicts  = [dict(zip(fields, d)) for d in data]
        series = [{ "name":"daily avg", "data": dicts }]
        title  = "Current year"
        xaxis  = "" #["Aug","","Sep","","Oct","","Nov","","Dec","","Jan","","Feb","","Mar","","Apr","","May","","Jun","","Jul",""]
    elif chartid == "global":
        series = [
            { "name":"2022-2023", "data": [ 0,  0, 100, 240, 330, 426, 421, 410, 180,  90,  0, 0 ] },
            { "name":"2021-2022", "data": [ 0, 50, 200, 340, 430, 526, 521, 610, 580, 290, 90, 0 ] },
            { "name":"2019-2020", "data": [ 0,  0, 100, 240, 330, 426, 421, 410, 180,  90,  0, 0 ] },
            ]
        title="Previous year consumption"
        xaxis = ["Aug","Sep","Oct","Nov","Dec","Jan","Feb","Mar","Apr","May","Jun","Jul"]

    return title, json.dumps(series, default=json_serial), json.dumps(xaxis)

@login_required
@app.route('/konsumo/chart/<prefix>', methods=['GET'])
def chart(prefix):
    title, series, xaxis = present_data(prefix)

    return render_template("chart.js", 
                    prefix=copy.copy(prefix), 
                    title=copy.copy(title),
                    series=copy.copy(series),
                    xaxis=copy.copy(xaxis),
                    )

@app.route("/konsumo/profile")
def profile():
    return render_template('profile.html')

@login_required
@app.route('/konsumo/location', methods=['GET','POST'])
def location():
    if request.method=='POST':
        location = request.form['location']
        user = User()
        user.set_location(current_user.id, location)
        return redirect(url_for('location'))
    return render_template('location.html')

@login_required
@app.route('/konsumo/form', methods=['GET','POST'])
def form():
    if request.method=='POST':
        date   = request.form['date']
        type   = request.form['type']
        value1 = request.form['value1']
        value2 = request.form['value2']+"" # ALLOWED NULL VALUE HERE

        # Convert date from DD-MM-YYYY to YYYY-MM-DD
        # date = datetime.strptime(date,'%d-%m-%Y').strftime('%Y-%m-%d')

        user = User()
        user.set_data(date, type, value1, value2, current_user.id)
        return redirect(url_for('form'))
    return render_template('form.html')

@app.route('/konsumo/', methods=['GET'])
@app.route('/konsumo', methods=['GET'])
def index():
    prefixes= [ "current", "global" ]
    return render_template("index.html", prefixes=copy.copy(prefixes), current_user=current_user)

@app.route('/')
def root():
    return redirect("/konsumo", code=302)

if __name__ == "__main__":
    # SSL Mode
    app.run(host=HOST, port=int(PORT), ssl_context="adhoc", debug=DEBUG)
    # No SSL (usage with gunicorn)
    # app.run(host=HOST, port=int(PORT), debug=DEBUG)
