from datetime import datetime, date, timedelta
import pandas as pd
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

# last_value = 0
# def diff_m(v):
#         global last_value
#         ret = v - last_value
#         last_value = v
#         return ret

# def construct_data(data):
#     global last_value
#     last_value = list(data)[0][1]
#     return [(k, diff_m(v)) for k,v in data ]

def construct_data(data, chart_type):
    df = pd.DataFrame(data)

    df.rename(columns = { df.columns[0]:'DATE', df.columns[1]:'CAPACITY' }, inplace=True)
    # Here we get the CAPACITY column and we shift it up, 
    # so the "CAPACITY" in a row will put in the previous row
    series_shifted = df['CAPACITY'].shift()

    # Here we create a "DIFF" column with the CAPACITY minus the previous value
    if chart_type == 'electricity':
        df['DIFF'] = df['CAPACITY'] - series_shifted
    elif chart_type == 'gazoline':
        df['DIFF'] = series_shifted - df['CAPACITY'] 

    # We remove the CAPACITY column
    df.drop('CAPACITY', inplace=True, axis=1)
    # Remove NaN
    df=df.dropna()
    # Remove .0
    df['DIFF'] = df['DIFF'].astype(int).astype(str)

    # For data construction purpose :
    df.rename(columns = { df.columns[0]:'x', df.columns[1]:'y' }, inplace=True)
    
    return df.to_dict('records')

def present_data(data, chartid):
    heating_period={ "start":"09", "end":"05" }

    chart_type = 'electricity'
    data = construct_data(data, chart_type)

    if chartid == "current":
        # fields = ['x', 'y']
        # dicts  = [dict(zip(fields, d)) for d in data]
        # print(dicts)
        series = [{ "name":"daily avg", "data": data }]
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