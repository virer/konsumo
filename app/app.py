#!/usr/bin/env python3
######################
# S.CAPS Mar 2023
# Konsumo
######################

# Logging level
DEBUG = True

from flask import Flask, render_template, redirect, url_for, request, jsonify, copy_current_request_context, abort, send_from_directory
from jinja2 import TemplateNotFound
from flask_login import LoginManager, current_user, login_required, login_user, logout_user 
from oauthlib.oauth2 import WebApplicationClient
import os, copy, json, requests, sqlite3
from user import User

app = Flask(__name__, static_url_path='/konsumo/static')
app.secret_key = os.environ.get("SECRET_KEY") or os.urandom(24)
app.config['SEND_FILE_MAX_AGE_DEFAULT'] = 0

# Configuration
HOST = os.getenv("HOST", "127.0.0.1")
PORT = os.getenv("PORT", "8080")
GOOGLE_CLIENT_ID = os.environ.get("GOOGLE_CLIENT_ID", None)
GOOGLE_CLIENT_SECRET = os.environ.get("GOOGLE_CLIENT_SECRET", None)
GOOGLE_DISCOVERY_URL = ("https://accounts.google.com/.well-known/openid-configuration")

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

@app.route("/konsumo/profile")
def profile():
    if current_user and current_user.is_authenticated:
        print(dir(current_user))
        return (
            "<p>Hello, {}! You're logged in! Email: {}</p>"
            "<div><p>Google Profile Picture:</p>"
            '<img src="{}" alt="Google profile pic"></img></div>'
            'Location: {}<br><br>'
            '<a class="button" href="/konsumo">content</a><br><br>'
            '<a class="button" href="/konsumo/logout">Logout</a>'.format(
                current_user.name, current_user.email, current_user.profile_pic, current_user.location
            )
        )
    else:
        return '<a class="button" href="/konsumo/login">Google Login</a>'

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

def construct_data(chartid):
    heating_period={ "start":"09", "end":"05" }

    if chartid == "current":
        series = [
            { "name":"daily avg", "data": [  0,  0,  2,  3,  6, 16, 17, 15,  10,  9,  0, 0, 30, 25, 16, 12,  8,  0,  2,  1,   3, 7, 17, 25 ] },
            { "name":"avg t°", "data":    [ 30, 25, 16, 12,  8,  0,  2,  1,   3, 7, 17, 25, 0,  0,  2,  3,  6, 16, 17, 15,  10,  9,  0, 0 ] }
            ]
        title="Current year"
        xaxis = ["Aug","","Sep","","Oct","","Nov","","Dec","","Jan","","Feb","","Mar","","Apr","","May","","Jun","","Jul",""]
    elif chartid == "global":
        series = [
            { "name":"2022-2023", "data": [ 0,  0, 100, 240, 330, 426, 421, 410, 180,  90,  0, 0 ] },
            { "name":"2021-2022", "data": [ 0, 50, 200, 340, 430, 526, 521, 610, 580, 290, 90, 0 ] },
            { "name":"2019-2020", "data": [ 0,  0, 100, 240, 330, 426, 421, 410, 180,  90,  0, 0 ] },
            ]
        title="Previous year consumption"
        xaxis = ["Aug","Sep","Oct","Nov","Dec","Jan","Feb","Mar","Apr","May","Jun","Jul"]

    return title, json.dumps(series), json.dumps(xaxis)

@login_required
@app.route('/konsumo/chart/<prefix>', methods=['GET'])
def chart(prefix):
    title, series, xaxis = construct_data(prefix)

    return render_template("chart.js", 
                    prefix=copy.copy(prefix), 
                    title=copy.copy(title),
                    series=copy.copy(series),
                    xaxis=copy.copy(xaxis),
                    )

@login_required
@app.route('/konsumo/location', methods=['GET','POST'])
def location():
    if request.method=='POST':
        location = request.form['location']
        user = User()
        user.set_location(current_user.id, location)
        return redirect(url_for('location'))
    return render_template('location.html')

@app.route('/konsumo/', methods=['GET'])
@app.route('/konsumo', methods=['GET'])
def index():
    prefixes= [ "current", "global" ]
    return render_template("index.html", prefixes=copy.copy(prefixes), current_user=current_user)

@app.route('/')
def root():
    return redirect("/konsumo", code=302)

if __name__ == "__main__":
    if DEBUG:
        user = User()
        print(user.get("117426397869268208059"))
    
    # SSL Mode
    app.run(host=HOST, port=int(PORT), ssl_context="adhoc", debug=DEBUG)
    # No SSL (usage with gunicorn)
    # app.run(host=HOST, port=int(PORT), debug=DEBUG)

