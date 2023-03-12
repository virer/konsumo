from flask import render_template, request, jsonify, abort
from flask_httpauth import HTTPBasicAuth
from konsumo.auth.models import User
from flask import Blueprint, request
bp = Blueprint('api', __name__, url_prefix='/konsumo/api')

auth = HTTPBasicAuth()

@auth.verify_password
def verify_password(username, password):
    try:
        location, access_key, secret_key = User.get_info(username)
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
    try:
        value2 = json_data['value2']+"" # ALLOWED NULL VALUE HERE
    except:
        value2 = ""

    User.set_data(date, type, value1, value2, auth.current_user())

    return jsonify(
                message=f"Data Saved",
                category="success",
                status=200
            )