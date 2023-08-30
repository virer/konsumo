from konsumo.auth.models import User
from datetime import datetime, date, timedelta
import pandas as pd
import calendar
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

def month_list(start, end):
    mylist = []
    date_list=calendar.month_abbr
    for i in range(len(date_list)):
        if i >= start:
            mylist.append(date_list[i])

    for i in range(len(date_list)):
        if i <= end and i != 0:
            mylist.append(date_list[i])

    mylist2 = []
    for i in range(len(mylist)):
        # mylist2.append("01-{}".format(mylist[i]))
        mylist2.append("15-{}".format(mylist[i]))
    
    return mylist2

def convert_date(dat):
    date_format = "%Y-%m-%d"
    return datetime.strptime(dat, date_format).date()

def construct_data(data, chart_type, no_transform=False):
    df = pd.DataFrame(data)
    # if DEBUG:
    #     print(df)

    if len(df.index) <= 0:
        if no_transform:
            return df
        return df.to_dict('records')

    if len(df.columns.values.tolist()) >= 3: 
        df.rename(columns = { df.columns[0]:'DATE', df.columns[1]:'CAPACITY', df.columns[2]:'VALUE2' }, inplace=True)
        if chart_type == 'electricity':
            df['TOTAL']= df['CAPACITY'] + df['VALUE2']
            df.drop('CAPACITY', inplace=True, axis=1)
            df.rename(columns = {'TOTAL':'CAPACITY'}, inplace = True)
        df.drop('VALUE2', inplace=True, axis=1)
    else:
        df.rename(columns = { df.columns[0]:'DATE', df.columns[1]:'CAPACITY' }, inplace=True)
    
    # if DEBUG:
    #     print(df)

    # Here we get the CAPACITY column and we shift it up, 
    # so the "CAPACITY" in a row will put in the previous row
    series_shifted = df['CAPACITY'].shift()

    # Here we create a "DIFF" column with the CAPACITY minus the previous value
    if chart_type == 'electricity' or chart_type == 'water' or chart_type == 'gaz' or chart_type == 'other_plus' :
        df['DIFF'] = df['CAPACITY'] - series_shifted
    elif chart_type == 'gazoline' or chart_type == 'gaz_tank' or chart_type == 'other_minus' :
        df['DIFF'] = series_shifted - df['CAPACITY']
        # Remove negative value:
        df = df[ (df['DIFF'] > 0) | (df['DIFF'].isnull())] 

    # We remove the CAPACITY column
    df.drop('CAPACITY', inplace=True, axis=1)
    # Remove NaN
    df=df.dropna()
    # Remove .0
    df['DIFF'] = df['DIFF'].astype(int).astype(str)

    if no_transform:
        df['DATE'] = pd.to_datetime(df['DATE'].astype(str), format='%Y-%m-%d') # This format is the good one , others with year in the end are buggy with df.loc[]
        df['DIFF'] = df['DIFF'].astype(int)
        return df
    else:
        # For data construction purpose for ApexCharts (it needs { x: <date>, y: <value>})
        df.rename(columns = { df.columns[0]:'x', df.columns[1]:'y' }, inplace=True)
    
        return df.to_dict('records')

def mean_avg_date_range(cal_list, i, year, large=False):
    day, month = cal_list[i].split('-')
    start = '{}-{}-{}'.format(year, month, day)
    if large:
        start = '{}-{}-{}'.format(year, month, "01")

    if month == 'Dec': 
        year = str(int(year)+1)
    day, month = cal_list[i+1].split('-')
    end   = '{}-{}-{}'.format(year, month, day)
    if large:
        end   = '{}-{}-{}'.format(year, month, "28")

    return start, end, year

def mean_avg(year, data, cal_list):
    # Fill NaN value with previous one
    data.fillna(method='ffill', inplace=True)
    # Fill NaN value with next one
    data.fillna(method='bfill', inplace=True)
    
    ret = []
    for i in range(len(cal_list)):
        if i+1 < len(cal_list):
            start, end, year = mean_avg_date_range(cal_list, i, year)

            df2 = data.loc[( data['DATE'] >= start ) & ( data['DATE'] < end )]
            a=round(df2['DIFF'].mean(), 2)
            
            # Select larger time frame to avoid null value by getting average/mean value in between
            if str(a) == "nan": 
                start, end, year = mean_avg_date_range(cal_list, i, year, large=True)
                df2 = data.loc[( data['DATE'] >= start ) & ( data['DATE'] < end )]
                a=round(df2['DIFF'].mean(), 2)
    
            # Replace Pandas null value(NaN) with null -> for Javascript
            if str(a) == "nan":                 
                a = "null"
            ret.append(a)
    
    return ret

def get_last_day_of_month(input, format='%Y-%m-%d'):
    dt=datetime.strptime(input, format)
    ret = calendar.monthrange(dt.year, dt.month)
    day = ret[1]
    # note:
    # ret [0]  = weekday of first day (between 0-6 ~ Mon-Sun))
    # ret [1] = last day of the month
    return day

def current_start_end_period(heating_period):
    m=date.today().month
    year=date.today().year
    start="{}-{:0>2}-01".format(date.today().year-1, m)
    lastday=get_last_day_of_month(start)
    end="{}-{:0>2}-{}".format(date.today().year, m, lastday)
    return start, end

    # TODO XXX FIX the heating period below : XXX FIXME TODO

    # # If we are after the heating period, display the previous period on the graph
    # if m < heating_period["start"]+3:
    #     year=year-1

    # # {}:0>2 } fill the string with leading Zero
    # start="{}-{:0>2}-01".format(date.today().year-1, heating_period["start"])
    # lastday=get_last_day_of_month(start)
    # end="{}-{:0>2}-{}".format(date.today().year, heating_period["end"], lastday)
    # return start, end

def get_heating_period():
    # FIXME : load this from user profile
    # if gazolie (but if elec this is non sens)
    # heating_period={ "start":9, "end":5 }
    heating_period={ "start": date.today().month, "end": (date.today().month + 11) % 12 or 12 }
    return heating_period


def present_data(user_id, chartid, chart_type):
    heating_period=get_heating_period()

    year = date.today().year 
    if chartid == "current":
        xaxis = 'type: "datetime "'
        title  = "Past 12 month consumption"
        start, end = current_start_end_period(heating_period)
        if DEBUG:
            print("start/end: {0}/{1}".format(start,end))

        data = User().get_data_period(user_id, chart_type, start, end, value2=True)
        data = construct_data(data, chart_type)
        series = [{ "name":"daily avg", "data": data }]
    elif chartid == "global":
        title="Previous years consumption"
        series = []

        cal_list = month_list(heating_period["start"], heating_period["end"])
        cal_list = list(dict.fromkeys(cal_list)) # remove duplicates
        xaxis = 'categories: {}'.format( cal_list )

        # List the 3 previous year
        lines = [ (str(year-3), str(year-2)), 
                  (str(year-2), str(year-1)), 
                  (str(year-1), str(year)) ]

        for y, next_year in lines:
            start = "{}-{:0>2}-01".format(y,         heating_period["start"])
            end   = "{}-{:0>2}-31".format(next_year, heating_period["end"])
            data   = User().get_data_period(user_id, chart_type, 
                                convert_date(start),
                                convert_date(end),
                                value2=True 
                            )
            if len(data) > 0:
                data = construct_data(data, chart_type, no_transform=True)
                data = mean_avg(y, data, cal_list)             
                series.append({ "name": "{}-{}".format(y,next_year), "data": data })

    return title, json.dumps(series, default=json_serial).replace('"null"','null'), xaxis
