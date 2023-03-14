from konsumo.auth.models import User
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

def construct_data(data, chart_type, no_transform=False):
    df = pd.DataFrame(data)

    df.rename(columns = { df.columns[0]:'DATE', df.columns[1]:'CAPACITY' }, inplace=True)
    # Here we get the CAPACITY column and we shift it up, 
    # so the "CAPACITY" in a row will put in the previous row
    series_shifted = df['CAPACITY'].shift()

    # Here we create a "DIFF" column with the CAPACITY minus the previous value
    if chart_type == 'electricity' or chart_type == 'water' or chart_type == 'gaz' or chart_type == 'other_plus' :
        df['DIFF'] = df['CAPACITY'] - series_shifted
    elif chart_type == 'gazoline' or chart_type == 'gaz-tank' or chart_type == 'other_minus' :
        df['DIFF'] = series_shifted - df['CAPACITY'] 

    # We remove the CAPACITY column
    df.drop('CAPACITY', inplace=True, axis=1)
    # Remove NaN
    df=df.dropna()
    # Remove .0
    df['DIFF'] = df['DIFF'].astype(int).astype(str)

    # Remove year from date column 
    # df['DATE'] = pd.to_datetime(df['DATE'].astype(str), format='%Y-%m-%d').dt.strftime('%d-%b')

    if no_transform:        
        return df['DIFF'].values.tolist(), df['DATE'].astype(str).values.tolist()
    else:
        # For data construction purpose :
        df.rename(columns = { df.columns[0]:'x', df.columns[1]:'y' }, inplace=True)
    
        return df.to_dict('records')

def convert_date(dat):
    date_format = "%Y-%m-%d"
    return datetime.strptime(dat, date_format).date()

def present_data(user_id, chartid, chart_type):
    # FIXME : load this from user profile
    heating_period={ "start":"09", "end":"05" }

    if chartid == "current":
        xaxis = 'type: "datetime "'
        title  = "Current year consumption"
        start = "2023-03-05"
        end = "2023-03-31"
        data = User().get_data_period(user_id, chart_type, start, end)
        data = construct_data(data, chart_type)
        series = [{ "name":"daily avg", "data": data }]
    elif chartid == "global":
        title="Previous year consumption"
        series = []
        lines = [ "2021", "2022", "2023"]
        for year in lines:
            next_year= str(int(year) + 1)
            start = year + "-" + heating_period["start"] + "-01"
            end   = next_year + "-" + heating_period["end"] + "-31"
            data   = User().get_data_period(user_id, chart_type, 
                                convert_date(start),
                                convert_date(end) 
                            )
            if len(data) > 0:
                date_list = []
                data, date_list = construct_data(data, chart_type, no_transform=True)
                print(data)                
                series.append({ "name": year+"-"+next_year, "data": data })

        xaxis = date_list
        xaxis = 'categories: {}'.format(xaxis)

    return title, json.dumps(series, default=json_serial), xaxis