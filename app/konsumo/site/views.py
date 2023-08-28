from flask import render_template, redirect, request, make_response, jsonify, copy_current_request_context, abort
from flask_login import current_user, login_required
from flask import Blueprint, redirect, request
from konsumo.auth.models import User
from .lib import *
import copy

bp = Blueprint('konsumo', __name__, url_prefix='/konsumo')

type_list = [ 'electricity', 'gaz', 'gaz_tank', 'gazoline', 'water', 'other_plus', 'other_minus' ]

@login_required
@bp.route('/regenerate-secret', methods=['POST'])
def regenerate():
    User.regenerate_key(current_user.id)
    return redirect('/konsumo/profile?notif=saved')

@bp.route('/profile', methods=['GET'])
def profile():
    chart_type = type_list[0]
    try:
        _, access_key, secret_key = User.get_info(current_user.id)
        notif_msg = 'saved' == request.args.get('notif') 
        return render_template('profile.html', chart_type=chart_type,
                               notif_msg=notif_msg, 
                               access_key=access_key, 
                               secret_key=secret_key)
    except Exception as e:
        print("Profile exception : {0}".format(e))
        return redirect('/konsumo/')

@login_required
@bp.route('/chart/<prefix>', methods=['GET'])
def chart(prefix):
    chart_type = request.args.get('type')
    if chart_type not in type_list: abort(400)

    title, series, xaxis = present_data(current_user.id, prefix, chart_type)

    resp = make_response(render_template('chart.js', 
                    chart_type=chart_type,
                    prefix=copy.copy(prefix), 
                    title=copy.copy(title),
                    series=copy.copy(series),
                    xaxis=copy.copy(xaxis),
                    ), 200)
    resp.headers['Content-Type'] = 'text/javascript'
    return resp

@login_required
@bp.route('/location', methods=['GET','POST'])
def location():
    chart_type = type_list[0]
    if request.method=='POST':
        location = request.form['location']
        User(current_user.id).set_location(current_user.id, location)
        return redirect('/konsumo/location?notif=saved')
    location, _, _ = User.get_info(current_user.id)
    notif_msg = 'saved' == request.args.get('notif')
    return render_template('location.html', 
                           chart_type=chart_type,
                           location=location, 
                           notif_msg=notif_msg)

@login_required
@bp.route('/encoding', methods=['GET'])
def encoding():
    global type_list
    chart_type = request.args.get('type')
    
    if chart_type == None:
        chart_type = type_list[0]
    if chart_type not in type_list: abort(400)

    notif_msg = 'saved' == request.args.get('notif')
    
    return render_template('encoding.html', 
                    type_list=type_list, 
                    chart_type=chart_type,                          
                    notif_msg=notif_msg)

@login_required
@bp.route('/data/add', methods=['POST'])
def data_add():
    global type_list

    chart_type = request.form['type']
    if chart_type == None:
            chart_type = type_list[0]

    if chart_type not in type_list: abort(400)
        
    date       = request.form['date']
    value1     = request.form['value1']
    value2     = request.form['value2']+"" # ALLOWED NULL VALUE HERE

    # Convert date from DD-MM-YYYY to YYYY-MM-DD
    # date = datetime.strptime(date,'%d-%m-%Y').strftime('%Y-%m-%d')

    User.set_data(date, chart_type, value1, value2, current_user.id)
    return redirect('/konsumo/encoding?notif=saved&type={0}'.format(chart_type))

@login_required
@bp.route('/data/list', methods=['GET'])
def data_list():
    global type_list
    chart_type = request.args.get('type')
    if chart_type == None:
        chart_type = type_list[0]

    if chart_type not in type_list: abort(400)
    data_list = User().get_raw_data(current_user.id, chart_type)

    return render_template('data_list.html', 
                    type_list=type_list,
                    chart_type=chart_type,
                    data_list=data_list)

@login_required
@bp.route('/data/del', methods=['POST'])
def data_del():
    chart_type = request.form['type']
    if chart_type not in type_list: abort(400)
    
    id = request.form['id']
    User().del_data(current_user.id, chart_type, id)

    return redirect('/konsumo/data/list?type={0}'.format(chart_type))

@bp.route('/charts', methods=['GET'])
def charts():
    global type_list
    prefixes= [ 'current', 'global' ]
    chart_type = request.args.get('type')
    if chart_type not in type_list: abort(400)

    return render_template('charts.html', 
            chart_type=chart_type, 
            type_list=type_list,
            prefixes=copy.copy(prefixes), 
            current_user=current_user)

@bp.route('/', methods=['GET'])
@bp.route('', methods=['GET'])
def index():
    chart_type = type_list[0]
    return render_template('index.html', 
                           chart_type=chart_type,
                           current_user=current_user)
