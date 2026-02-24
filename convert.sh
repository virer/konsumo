#!/bin/bash

# cat /tmp/elec2.csv | sed -e 's/\([0-9][0-9]\)-\([0-9][0-9]\)-\([0-9][0-9][0-9][0-9]\);\(.*\);\(.*\);/{ "date": "\3-\2-\1T00:00:00Z", "category": "electricity", "electricity_day": \4, "electricity_night": \5 },/g'


cat /tmp/mazout.csv | sed -e 's/\(.*\);\([0-9][0-9]\)-\([0-9][0-9]\)-\([0-9][0-9]\)$/{ "date": "\4-\3-\2T00:00:00Z", "category": "fuel", "fuel": \1 },/g'

