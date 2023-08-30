from datetime import datetime
import pandas as pd


df = pd.read_csv('gazoline.csv',sep=';', names=["value1","date"])
df['date'] = pd.to_datetime(df['date'], format='%d-%m-%y')
df['date'] = df['date'].dt.strftime('%Y-%m-%d')
df.set_index('date')
# datetime.strptime(mydate, "%Y-%m-%d")

with open('gazoline.json', 'w', encoding='utf-8') as file:
    df.to_json(file, force_ascii=False, orient='records', index=True, indent=4)

