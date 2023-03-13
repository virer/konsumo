from flask import render_template, redirect, request, jsonify, copy_current_request_context, abort
from flask_login import current_user, login_required
from flask import Blueprint, redirect, request
from konsumo.auth.models import User
from .lib import *
import copy

bp = Blueprint('konsumo', __name__, url_prefix='/konsumo')

@login_required
@bp.route('/regenerate-secret', methods=['POST'])
def regenerate():
    User.regenerate_key(current_user.id)
    return redirect('/konsumo/profile?notif=saved')

@bp.route('/profile', methods=['GET'])
def profile():
    location, access_key, secret_key = User.get_info(current_user.id)
    notif_msg = 'saved' == request.args.get('notif') 
    return render_template('profile.html', notif_msg=notif_msg, access_key=access_key, secret_key=secret_key)

@login_required
@bp.route('/chart/<prefix>', methods=['GET'])
def chart(prefix):

    # elec FIXME TODO
    type = 'electricity'
    # Get data here
    data = User().get_data(current_user.id, type)

    title, series, xaxis = present_data(data, prefix)

    return render_template('chart.js', 
                    prefix=copy.copy(prefix), 
                    title=copy.copy(title),
                    series=copy.copy(series),
                    xaxis=copy.copy(xaxis),
                    )

@login_required
@bp.route('/location', methods=['GET','POST'])
def location():
    if request.method=='POST':
        location = request.form['location']
        User(current_user.id).set_location(current_user.id, location)
        return redirect('/konsumo/location?notif=saved')
    location, access_key, secret_key = User.get_info(current_user.id)
    notif_msg = 'saved' == request.args.get('notif')
    return render_template('location.html', location=location, notif_msg=notif_msg)

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

        User.set_data(date, type, value1, value2, current_user.id)
        return redirect('/konsumo/form?notif=saved')        
    notif_msg = 'saved' == request.args.get('notif')
    return render_template('form.html', notif_msg=notif_msg)

@bp.route('/charts', methods=['GET'])
def charts():
    prefixes= [ 'current', 'global' ]
    return render_template('charts.html', prefixes=copy.copy(prefixes), current_user=current_user)

@bp.route('/', methods=['GET'])
@bp.route('', methods=['GET'])
def index():
    return render_template('index.html', current_user=current_user)
