from datetime import datetime, date, timedelta
import json

DEBUG = True

def json_serial(obj):
    """JSON serializer for objects not serializable by default json code"""

    if isinstance(obj, (datetime, date)):
        return obj.isoformat()
    raise TypeError ("Type %s not serializable" % type(obj))

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

def construct_data(data):
    global last_value
    last_value = list(data)[0][1]

    return [(k, diff_m(v)) for k,v in data ]

def present_data(data, chartid):
    heating_period={ "start":"09", "end":"05" }

    data = construct_data(data)

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