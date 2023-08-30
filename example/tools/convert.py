from datetime import datetime
import pandas as pd

# Examples
# file_in='gazoline.csv'
# column_order=["value1","date"]
# separator=';'

# file_in = 'water.csv'
# column_order = ["date","value1"]
# separator = "\t" # Tab or ';' or ','
# file_out = 'water.json'

file_in = 'elec.csv'
column_order = ["date","value1","value2"]
separator = "\t" # Tab or ';' or ','
file_out = 'elec.json'

# ********************
df = pd.read_csv(file_in, sep=separator, names=column_order)

df['date'] = pd.to_datetime(df['date'], format='%d-%m-%y')
df['date'] = df['date'].dt.strftime('%Y-%m-%d')
df.set_index('date')

with open(file_out, 'w', encoding='utf-8') as file:
    df.to_json(file, force_ascii=False, orient='records', index=True, indent=4)

