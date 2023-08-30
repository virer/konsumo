from flask import render_template, request, jsonify, abort
from flask_httpauth import HTTPBasicAuth
from konsumo.auth.models import User
from flask import Blueprint, request

bp = Blueprint('api', __name__, url_prefix='/konsumo/api')

auth = HTTPBasicAuth()

@auth.verify_password
def verify_password(username, password):
    try:
        if username.isdigit():
            _, _, secret_key = User.get_info(username)
            if password == secret_key:
                return username
    except:
        pass
    return False

@bp.route('/add/<type>', methods=['POST'])
@auth.login_required
def api_add_one(type):
    if request.headers['Content-Type'] != 'application/json':
        bp.logger.debug(request.headers['Content-Type'])
        return jsonify(msg=('Header Error'))

    json_data = request.get_json()
    if not json_data:
            return {"message": "Json data not present"}, 400

    date   = json_data["date"]
    value1 = json_data['value1']
    value2 = ""
    if 'value2' in json_data: # Allowed null value2
        value2 = str(json_data['value2'])

    User.set_data(date, type, value1, value2, auth.current_user())

    return jsonify(
                message="Data Saved",
                status="success",
            )

@bp.route('/addbundle/<type>', methods=['POST'])
@auth.login_required
def api_add_multi(type):
    if request.headers['Content-Type'] != 'application/json':
        bp.logger.debug(request.headers['Content-Type'])
        return jsonify(msg=('Header Error'))

    json_datas = request.get_json()
    if not json_datas:
            return {"message": "Json data not present"}, 400

    for json_data in json_datas:
        date   = json_data["date"]
        value1 = json_data['value1']
        value2 = ""
        if 'value2' in json_data: # Allowed null value2
            value2 = str(json_data['value2'])

        User.set_data(date, type, value1, value2, auth.current_user())

    return jsonify(
                message="Data Saved",
                status="success",
            )