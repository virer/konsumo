from flask import Flask, render_template, redirect, url_for, request, jsonify, copy_current_request_context, abort
from flask_login import current_user, login_required
from flask import Blueprint, redirect, url_for, request
from konsumo.auth.models import User
from .lib import *
import copy

bp = Blueprint("konsumo", __name__, url_prefix="/konsumo")

@login_required
@bp.route('/chart/<prefix>', methods=['GET'])
def chart(prefix):

    # elec FIXME TODO
    type = 'electricity'
    # Get data here
    user = User()
    data = user.get_data(current_user.id, type)

    title, series, xaxis = present_data(data, prefix)

    return render_template("chart.js", 
                    prefix=copy.copy(prefix), 
                    title=copy.copy(title),
                    series=copy.copy(series),
                    xaxis=copy.copy(xaxis),
                    )

@bp.route("/profile")
def profile():
    return render_template('profile.html')

@login_required
@bp.route('/location', methods=['GET','POST'])
def location():
    if request.method=='POST':
        location = request.form['location']
        user = User()
        user.set_location(current_user.id, location)
        return redirect(url_for('location'))
    return render_template('location.html')

@login_required
@bp.route('/form', methods=['GET','POST'])
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

@bp.route('/', methods=['GET'])
@bp.route('', methods=['GET'])
def index():
    prefixes= [ "current", "global" ]
    return render_template("index.html", prefixes=copy.copy(prefixes), current_user=current_user)

# @bp.route('/')
# def root():
#     user_id = "117426397869268208059"
#     type = "water"
#     user = User()
#     print("TEST 1***")
#     data = user.get(user_id)
#     print(data)
#     print("TEST 2***")
#     _date = "2023-03-11"
#     value1 = 356
#     value2 = ""
#     data = user.set_data(_date, type, value1, value2, user_id)
#     print(data)
#     print("TEST 3***")
#     return redirect("/konsumo", code=302)